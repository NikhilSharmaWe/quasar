package router

import (
	"io"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/websocket"
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

// func (app *Application) sendOldChats(peerConnection) error {
// 	filter := bson.M{
// 		"meeting_key": meetingKey,
// 	}

// 	chats, err := app.ChatRepo.Find(filter)
// 	if err != nil {
// 		return err
// 	}

// 	for _, chat := range chats {

// 	}

// }

func (app *Application) messageClients(msg *model.Chat) error {
	for _, peerConnection := range app.PeerConnections {
		if peerConnection.Key == msg.MeetingKey {
			err := app.messageClient(peerConnection.Websocket.Conn, msg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) messageClient(ws *websocket.Conn, message *model.Chat) error {
	err := ws.WriteJSON(message)
	if err != nil && unsafeError(err) {
		ws.Close()
		return err
	}
	return nil
}

func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}
