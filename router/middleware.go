package router

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	echo "github.com/labstack/echo/v4"
)

func (app *Application) createSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := app.CookieStore.Get(c.Request(), "signin") // this will create the cookie if it does not exists
		if err != nil {
			return err
		}

		c.Set("session", session)
		return next(c)
	}
}

func (app *Application) alreadyLoggedIn(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(*sessions.Session)

		authenticated, ok := session.Values["authenticated"].(bool)
		fmt.Println(authenticated)
		if ok && authenticated {
			return c.Redirect(http.StatusFound, "/meets/")
		}

		return next(c)
	}
}
