package router

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type PeerConnectionState struct {
	*webrtc.PeerConnection
	Websocket *ThreadSafeWriter
}

type ThreadSafeWriter struct {
	Lock sync.Mutex
	Conn *websocket.Conn
}

// func (app *Application) SignalPeerConnections() {
// 	app.Lock()
// 	defer func() {
// 		app.Unlock()
// 	}()

// 	attemptSync := func() (tryAgain bool) {
// 		for i := range app.PeerConnections {}
// 		return
// 	}
// }

func (app *Application) dispatchKeyFrame() {
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

func (t *ThreadSafeWriter) writeJSON(v interface{}) error {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	return t.Conn.WriteJSON(v)
}
