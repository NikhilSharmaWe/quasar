package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	alreadyExistsErr = "element already exists"
)

func convertToBSON(m Model) (*bson.M, error) {
	b, err := bson.Marshal(m)
	if err != nil {
		return nil, err
	}

	doc := &bson.M{}
	err = bson.Unmarshal(b, doc)

	return doc, err
}

func userFromForm(c echo.Context) *User {
	user := &User{
		Username:  c.FormValue("username"),
		Firstname: c.FormValue("firstname"),
		Lastname:  c.FormValue("lastname"),
		Password:  []byte(c.FormValue("password")),
	}
	return user
}

func setSession(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.ID = uuid.NewV4().String()
	session.Values["username"] = c.FormValue("username")
	return session.Save(c.Request(), c.Response())
}

func getSessionHandler(c echo.Context) *sessions.Session {
	session := c.Get("session").(*sessions.Session)
	return session
}

func clearSessionHandler(c echo.Context) error {
	session := c.Get("session").(*sessions.Session)
	session.Options.MaxAge = -1
	session.Save(c.Request(), c.Response())
	return c.String(http.StatusOK, "Session Cleared")
}

func createSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := store.Get(c.Request(), "login")
		if err != nil {
			return err
		}

		c.Set("session", session)
		return next(c)
	}
}

// make this the middleware instead of createsessionMiddleware. not sure
func alreadyLoggedIn() {

}
