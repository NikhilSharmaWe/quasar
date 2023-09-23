package router

import (
	"fmt"
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
		if err := app.messageClient(pcState, &chat); err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) messageClients(chat *model.Chat) error {
	for _, pcState := range app.PeerConnections {
		if pcState.Key == chat.MeetingKey {
			err := app.messageClient(pcState, chat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) messageClient(pcState PeerConnectionState, chat *model.Chat) error {
	fmt.Println(*chat)
	err := pcState.Websocket.WriteJSON(WebsocketMessage{
		Event: "chat",
		Chat:  *chat,
	})
	fmt.Printf("sending message to %s", pcState.Username)
	if err != nil && unsafeError(err) {
		pcState.Websocket.Conn.Close()
		return err
	}
	return nil
}

func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}
