package router

import (
	"io"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

func (app *Application) HandleMessages() {
	for {
		msg := <-app.Broadcaster
		if err := app.ChatRepo.Save(msg); err != nil {
			app.Logger.Println(err)
		}

		if err := app.messageClients(msg); err != nil {
			app.Logger.Println(err)
		}
	}
}

func (app *Application) sendOldChats(pcState PeerConnectionState) error {
	filter := bson.M{
		"meeting_key": pcState.Key,
	}

	chats, err := app.ChatRepo.Find(filter)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		if err := app.messageClientChat(pcState, &chat); err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) messageClients(chat *model.Chat) error {
	for _, pcState := range app.PeerConnections {
		if pcState.Key == chat.MeetingKey {
			err := app.messageClientChat(pcState, chat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) messageClientChat(pcState PeerConnectionState, data *model.Chat) error {
	err := pcState.Websocket.WriteJSON(WebsocketMessage{
		Event: "chat",
		Data:  &data,
	})

	if err != nil && unsafeError(err) {
		pcState.Websocket.Conn.Close()
		return err
	}
	return nil
}

func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

type ParticipantInfo struct {
	StreamID string
	Username string
}

func (app *Application) messageClientsRemoteUserInfo(latestStream string) error {
	newParticipantStreamID := latestStream
	newParticipantUsername := app.StreamInfo[latestStream]

	for _, pcState := range app.PeerConnections {
		// send the new participant info of all other participants in the meeting
		if pcState.Username == newParticipantUsername {
			newParticipantPCState := pcState
			for streamID, username := range app.StreamInfo {
				if username == newParticipantUsername {
					continue
				}
				if err := newParticipantPCState.Websocket.WriteJSON(WebsocketMessage{
					Event: "participant",
					Data: ParticipantInfo{
						StreamID: streamID,
						Username: username,
					},
				}); err != nil {
					return err
				}
			}
			continue
		}

		// send all other participants info of the new one
		if err := pcState.Websocket.WriteJSON(WebsocketMessage{
			Event: "participant",
			Data: ParticipantInfo{
				StreamID: newParticipantStreamID,
				Username: newParticipantUsername,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
