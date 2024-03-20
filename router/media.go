package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/kluctl/go-embed-python/python"
	echo "github.com/labstack/echo/v4"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type PeerConnectionState struct {
	Key      string
	Username string
	*webrtc.PeerConnection
	Websocket *ThreadSafeWriter
}

type ThreadSafeWriter struct {
	sync.Mutex
	*websocket.Conn
}

func (app *Application) WebsocketHandler(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)

	username := session.Values["username"].(string)
	meetingKey := session.Values["meeting_key"].(string)

	unsafeConn, err := app.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	conn := &ThreadSafeWriter{sync.Mutex{}, unsafeConn}

	err = conn.WriteJSON(WebsocketMessage{
		Event: "my_username",
		Data:  username,
	})

	if err != nil {
		c.Logger().Error(err)
		return err
	}

	defer conn.Close()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	})
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	defer peerConnection.Close()

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			c.Logger().Error(err)
			return err
		}
	}

	pcState := PeerConnectionState{
		Key:            meetingKey,
		Username:       username,
		PeerConnection: peerConnection,
		Websocket:      conn,
	}

	app.Lock()
	app.PeerConnections = append(app.PeerConnections, pcState)
	app.Unlock()

	app.configurePCEvents(username, meetingKey, &pcState)
	app.signalPeerConnections()

	app.sendOldChats(pcState)

	message := &WebsocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			c.Logger().Error(err)
			return err
		} else if err := json.Unmarshal(raw, &message); err != nil {
			c.Logger().Error(err)
			return err
		}

		switch message.Event {
		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			data, ok := message.Data.(string)
			if !ok {
				c.Logger().Error("unable to parse message data")
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}
			if err := json.Unmarshal([]byte(data), &candidate); err != nil {
				c.Logger().Error(err)
				return err
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				c.Logger().Error(err)
				return err
			}

		case "answer":
			answer := webrtc.SessionDescription{}
			data, ok := message.Data.(string)
			if !ok {
				c.Logger().Error("unable to parse message data")
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}
			if err := json.Unmarshal([]byte(data), &answer); err != nil {
				c.Logger().Error(err)
				return err
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				c.Logger().Error(err)
				return err
			}

		case "chat":
			data, ok := message.Data.(string)
			if !ok {
				c.Logger().Error("unable to parse message data")
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}
			chat := model.Chat{
				MeetingKey: pcState.Key,
				Username:   pcState.Username,
				Message:    data,
			}

			app.Broadcaster <- &chat

		case "code":
			data, ok := message.Data.(string)
			if !ok {
				c.Logger().Error("unable to parse message data")
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}

			code := model.Code{
				MeetingKey: pcState.Key,
				Code:       data,
			}

			app.Broadcaster <- &code

			fmt.Println("CODE:", data)

		case "compile":
			data, ok := message.Data.(string)
			if !ok {
				c.Logger().Error("unable to parse message data")
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}

			result := data[1 : len(data)-1]
			newStr := strings.Replace(result, "\\n", "\n", -1)
			newStr = strings.ReplaceAll(newStr, "\\", "")
			// fmt.Println("COMPILE:", data)
			compilePythonAndSend(newStr, conn)
		}
	}
}

func compilePythonAndSend(code string, conn *ThreadSafeWriter) {
	ep, err := python.NewEmbeddedPython("example")
	if err != nil {
		panic(err)
	}
	fmt.Println("")
	cmd := ep.PythonCmd("-c", code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	var output string
	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		// If there's an error, print stderr
		// fmt.Println("Error:", stderr.String())
		output = stderr.String()
		// fmt.Println("Stderr:", stderr.String())
	} else {
		// If command completes successfully, print stdout
		// fmt.Println("Output:", stdout.String())
		output = stdout.String()
	}

	if err = conn.WriteJSON(&WebsocketMessage{
		Event: "output",
		Data:  output,
	}); err != nil {
		log.Println("Error:", err)
	}

}

func (app *Application) signalPeerConnections() {
	app.Lock()
	defer func() {
		app.Unlock()
		app.DispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		fmt.Println("Hello")
		for i := range app.PeerConnections {
			if app.PeerConnections[i].ConnectionState() == webrtc.PeerConnectionStateClosed {
				app.PeerConnections = append(app.PeerConnections[:i], app.PeerConnections[i+1:]...)
				return true
			}

			existingSenders := map[string]bool{}

			for _, sender := range app.PeerConnections[i].GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				if _, ok := app.TrackLocals[sender.Track().ID()]; !ok {
					if err := app.PeerConnections[i].RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			for _, receiver := range app.PeerConnections[i].GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			for trackID := range app.TrackLocals {
				fmt.Printf("tracklocal username: %s and peerConnection username: %s", app.TrackLocals[trackID].Username, app.PeerConnections[i].Username)
				if app.TrackLocals[trackID].MeetingKey == app.PeerConnections[i].Key {
					// && (app.TrackLocals[trackID].Username != app.PeerConnections[i].Username) {
					if _, ok := existingSenders[trackID]; !ok {
						if _, err := app.PeerConnections[i].AddTrack(app.TrackLocals[trackID]); err != nil {
							return true
						}
					}
				}
			}

			offer, err := app.PeerConnections[i].CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = app.PeerConnections[i].SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = app.PeerConnections[i].Websocket.WriteJSON(&WebsocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}
		return false
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				app.signalPeerConnections()
			}()

			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (app *Application) addTrack(t *webrtc.TrackRemote, meetingKey, username string) (*webrtc.TrackLocalStaticRTP, error) {
	app.Lock()
	defer func() {
		app.Unlock()
		app.signalPeerConnections()
	}()

	app.StreamInfo[t.StreamID()] = username

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		return nil, err
	}

	app.TrackLocals[t.ID()] = TrackLocal{
		MeetingKey:          meetingKey,
		Username:            username,
		TrackLocalStaticRTP: trackLocal,
	}

	return trackLocal, nil
}

func (app *Application) removeTrack(t *webrtc.TrackLocalStaticRTP) {
	app.Lock()
	defer func() {
		app.Unlock()
		app.signalPeerConnections()
	}()

	delete(app.TrackLocals, t.ID())
	delete(app.StreamInfo, t.StreamID())
}

func (app *Application) DispatchKeyFrame() {
	app.Lock()
	defer app.Unlock()

	for _, peerConnection := range app.PeerConnections {
		for _, reciever := range peerConnection.GetReceivers() {
			if reciever.Track() == nil {
				continue
			}

			_ = peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(reciever.Track().SSRC()),
				},
			})
		}
	}
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}

func (app *Application) configurePCEvents(username, meetingKey string, pcState *PeerConnectionState) {
	pcState.PeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}

		if writeErr := pcState.Websocket.WriteJSON(&WebsocketMessage{
			Event: "candidate",
			Data:  string(candidateString),
		}); writeErr != nil {
			log.Println(writeErr)
		}
	})

	pcState.PeerConnection.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		switch pcs {
		case webrtc.PeerConnectionStateFailed:
			if err := pcState.PeerConnection.Close(); err != nil {
				log.Println(err)
			}
		case webrtc.PeerConnectionStateClosed:
			app.signalPeerConnections()
		}
	})

	pcState.PeerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		trackLocal, err := app.addTrack(tr, pcState.Key, pcState.Username)
		if err != nil {
			return
		}

		app.messageClientsRemoteUserInfo(trackLocal.StreamID())

		defer app.removeTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := tr.Read(buf)
			if err != nil {
				return
			}

			if _, err := trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})
}
