package router

import (
	"net/http"

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

func (app *Application) ifAlreadyLogined(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if alreadyLoggedIn(c) {
			return c.Redirect(http.StatusFound, "/meets")
		}
		return next(c)
	}
}

func (app *Application) ifNotLogined(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !alreadyLoggedIn(c) {
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}
