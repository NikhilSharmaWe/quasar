package router

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	Context     context.Context
	Logger      *log.Logger
	UserRepo    *model.GenericRepo[model.User]
	MeetingRepo *model.GenericRepo[model.Meeting]
	ChatRepo    *model.GenericRepo[model.Chat]
	CodeRepo    *model.GenericRepo[model.Code]
	CookieStore *sessions.CookieStore
	sync.RWMutex
	PeerConnections []PeerConnectionState
	TrackLocals     map[string]TrackLocal
	StreamInfo      map[string]string
	websocket.Upgrader
	Broadcaster chan interface{}
}

type TrackLocal struct {
	MeetingKey string
	Username   string
	*webrtc.TrackLocalStaticRTP
}

func NewApplication() *Application {
	var (
		ctx      = context.Background()
		logger   = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		userRepo = model.GenericRepo[model.User]{
			Collection: "user",
		}
		meetingRepo = model.GenericRepo[model.Meeting]{
			Collection: "meeting",
		}
		chatRepo = model.GenericRepo[model.Chat]{
			Collection: "chat",
		}
		codeRepo = model.GenericRepo[model.Code]{
			Collection: "code",
		}
		cookieStore     = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))
		peerConnections = []PeerConnectionState{}
		trackLocals     = make(map[string]TrackLocal)
		streamInfo      = make(map[string]string)
		upgrader        = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		broadcaster = make(chan interface{})
	)

	application := &Application{
		Context:         ctx,
		Logger:          logger,
		UserRepo:        &userRepo,
		MeetingRepo:     &meetingRepo,
		ChatRepo:        &chatRepo,
		CodeRepo:        &codeRepo,
		CookieStore:     cookieStore,
		PeerConnections: peerConnections,
		TrackLocals:     trackLocals,
		StreamInfo:      streamInfo,
		RWMutex:         sync.RWMutex{},
		Upgrader:        upgrader,
		Broadcaster:     broadcaster,
	}

	return application
}

func setSession(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.ID = uuid.NewV4().String()
	session.Values["username"] = c.FormValue("username")
	session.Values["authenticated"] = true
	return session.Save(c.Request(), c.Response())
}

func clearSessionHandler(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.Options.MaxAge = -1
	return session.Save(c.Request(), c.Response())
}

func userFromContext(c echo.Context) (*model.User, error) {
	var user *model.User
	bs, err := bcrypt.GenerateFromPassword([]byte(c.FormValue("password")), bcrypt.MinCost)
	if err != nil {
		return user, err
	}

	user = &model.User{
		Username:  c.FormValue("username"),
		Firstname: c.FormValue("firstname"),
		Lastname:  c.FormValue("lastname"),
		Password:  bs,
	}
	return user, nil
}

func meetingFromContext(c echo.Context) (*model.Meeting, error) {
	session := c.Get("session").(*sessions.Session)
	un := session.Values["username"].(string)
	meetingKey := uuid.NewV4().String()

	meeting := &model.Meeting{
		Organizer:  un,
		MeetingKey: meetingKey,
	}

	return meeting, nil
}

func (app *Application) alreadyLoggedIn(c echo.Context) bool {
	session := c.Get("session").(*sessions.Session)

	username, ok := session.Values["username"].(string)
	if !ok {
		return false
	}

	filter := make(map[string]interface{})
	filter["username"] = username

	exists, err := app.UserRepo.IsExists(filter)
	if err != nil {
		return false
	}

	if !exists {
		return false
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	if ok && authenticated {
		return true
	}

	return false
}
