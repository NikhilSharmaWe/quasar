package router

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/sessions"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) Router() *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(app.createSessionMiddleware)

	e.Static("/assets", "./public")

	e.GET("/", ServeHTML("./public/login/index.html"), app.IfAlreadyLogined)
	e.GET("/signup", ServeHTML("./public/signup/index.html"), app.IfAlreadyLogined)
	e.GET("/meets", ServeHTML("./public/meets/index.html"), app.IfNotLogined)
	e.GET("/logout", app.HandleLogout)
	e.GET("/meets/:key", app.HandleMeeting, app.IfNotLogined)
	e.GET("/websocket", app.WebsocketHandler, app.IfNotLogined)

	e.POST("/", app.HandleSignIn)
	e.POST("/signup", app.HandleSignUp)
	e.POST("/create-meeting", app.HandleCreateMeeting, app.IfNotLogined)
	e.POST("/join-meeting", app.HandleJoinMeeting, app.IfNotLogined)

	return e
}

func ServeHTML(htmlPath string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(htmlPath)
	}
}

func (app *Application) HandleSignUp(c echo.Context) error {
	user, err := userFromContext(c)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	user.CreatedAt = time.Now()

	if err := app.UserRepo.SaveAccount(user); err != nil {
		if err.Error() == model.AlreadyExistsErr {
			return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
		}

		c.Logger().Error(err)
		return err
	}

	if err := setSession(c); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := c.Redirect(http.StatusSeeOther, "/meets/"); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleSignIn(c echo.Context) error {
	username := c.FormValue("username")
	user, err := app.UserRepo.FindOne(bson.M{"username": username})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user not found")
	}

	password := c.FormValue("password")
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "wrong password")
	}

	if err := setSession(c); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := c.Redirect(http.StatusSeeOther, "/meets/"); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleLogout(c echo.Context) error {
	if err := clearSessionHandler(c); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := c.Redirect(http.StatusSeeOther, "/"); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleCreateMeeting(c echo.Context) error {
	meeting, err := meetingFromContext(c)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	meeting.CreatedAt = time.Now()

	originalKey := meeting.MeetingKey
	encodedKey := make([]byte, base64.StdEncoding.EncodedLen(len([]byte(originalKey))))
	base64.StdEncoding.Encode(encodedKey, []byte(originalKey))

	meeting.MeetingKey = string(encodedKey)
	if err := app.MeetingRepo.Save(meeting); err != nil {
		c.Logger().Error(err)
		return err
	}

	if err := c.Redirect(http.StatusSeeOther, fmt.Sprintf("/meets/%s", originalKey)); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (app *Application) HandleJoinMeeting(c echo.Context) error {
	meetingKey := c.FormValue("meeting-id")
	lastIndex := strings.LastIndex(meetingKey, "/")

	if lastIndex != -1 {
		meetingKey = meetingKey[lastIndex+1:]
	}

	originalKey := meetingKey
	encodedKey := make([]byte, base64.StdEncoding.EncodedLen(len([]byte(originalKey))))
	base64.StdEncoding.Encode(encodedKey, []byte(originalKey))

	filter := make(map[string]interface{})
	filter["meeting_key"] = string(encodedKey)

	exists, err := app.MeetingRepo.IsExists(filter)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	if !exists {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid meeting key")
	}

	if err := c.Redirect(http.StatusSeeOther, fmt.Sprintf("/meets/%s", meetingKey)); err != nil {
		c.Logger().Error(err)
		return nil
	}

	return nil
}

func (app *Application) HandleMeeting(c echo.Context) error {
	meetingKey := c.Param("key")

	session := c.Get("session").(*sessions.Session)
	session.ID = uuid.NewV4().String()
	session.Values["meeting_key"] = meetingKey
	session.Save(c.Request(), c.Response())

	if err := c.File("./public/meeting/index.html"); err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}
