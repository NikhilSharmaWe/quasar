package router

import (
	"fmt"
	"io"
	"reflect"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

func (app *Application) HandleMessages() {
	for {
		msg := <-app.Broadcaster
		if reflect.TypeOf(&model.Chat{}) == reflect.TypeOf(msg) {
			fmt.Println("Hello")
			m, ok := msg.(*model.Chat)
			if !ok {
				app.Logger.Println("wrong message type")
			}

			if err := app.ChatRepo.Save(m); err != nil {
				app.Logger.Println(err)
			}

			if err := app.messageClients(m); err != nil {
				app.Logger.Println(err)
			}
		} else {
			c, ok := msg.(*model.Code)

			if !ok {
				app.Logger.Println("wrong message type")
			}

			exists, err := app.CodeRepo.IsExists(map[string]interface{}{"meeting_key": c.MeetingKey})
			if err != nil {
				app.Logger.Println(err)
			}

			if exists {
				// docID := c.ID
				// objID, err := primitive.ObjectIDFromHex(docID)
				// if err != nil {
				// 	app.Logger.Println(err)
				// 	fmt.Println("888888888888")
				// }

				filter := bson.M{"meeting_key": bson.M{"$eq": c.MeetingKey}}
				// filter := bson.D{
				// 	{"$and",
				// 		bson.A{
				// 			bson.D{
				// 				{"_id", bson.D{{"$gt", 25}}},
				// 			},
				// 		},
				// 	},
				// }
				update := bson.D{
					{"$set",
						bson.D{
							{"code", c.Code},
						},
					},
				}

				// id, _ := primitive.ObjectIDFromHex(c.ID)
				fmt.Println("----------")
				// app.CodeRepo.UpdateOne(map[string]interface{}{"meeting_key": c.MeetingKey}, c)
				// if err := app.CodeRepo.UpdateOne(map[string]interface{}{"_id": id}, c); err != nil {
				if err := app.CodeRepo.UpdateOne(filter, update); err != nil {

					fmt.Println("vvvvvvvvvv")
					app.Logger.Println(err)
				}
			} else {
				fmt.Println("//////")
				if err := app.CodeRepo.Save(c); err != nil {
					fmt.Println("xxxxxxxxxx")
					app.Logger.Println(err)
				}
			}

			if err := app.messageClientsCode(c); err != nil {
				app.Logger.Println(err)
			}

		}

		// fmt.Println("Hello")
		// m, ok := msg.(*model.Chat)
		// if !ok {
		// 	fmt.Println("HEYYYYYY")
		// }
		// if err := app.ChatRepo.Save(m); err != nil {
		// 	app.Logger.Println(err)
		// }

		// if err := app.messageClients(m); err != nil {
		// 	app.Logger.Println(err)
		// }

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

func (app *Application) messageClientsCode(code *model.Code) error {
	for _, pcState := range app.PeerConnections {
		if pcState.Key == code.MeetingKey {
			err := app.messageClientCode(pcState, code)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) messageClientCode(pcState PeerConnectionState, data *model.Code) error {
	err := pcState.Websocket.WriteJSON(WebsocketMessage{
		Event: "code",
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
				// if username == newParticipantUsername {
				// 	continue
				// }
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
