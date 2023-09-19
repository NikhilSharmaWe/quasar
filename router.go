package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	echo "github.com/labstack/echo/v4"
)

var (
	store = sessions.NewCookieStore([]byte(SESSION_SEC_KEY))
)

func (app *application) router() *echo.Echo {
	e := echo.New()

	e.Use(createSessionMiddleware)

	e.Static("/", "./public/login")
	e.Static("/signup/", "./public/signup")
	e.Static("/meets/", "./public/meets")

	e.POST("/signup/", app.HandleSignup)
	e.GET("/usermeet/", app.HandleMeet)
	return e
}

func (app *application) HandleSignup(c echo.Context) error {
	user := userFromForm(c)
	err := app.userRepo.SaveAccount(user)
	if err != nil {
		if err.Error() == alreadyExistsErr {
			return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
		}

		app.logger.Println("error:", err)
		return err
	}

	session := getSessionHandler(c)
	fmt.Printf("%+v", session)

	return c.Redirect(http.StatusSeeOther, "/meets/")
}

func (app *application) HandleMeet(c echo.Context) error {
	session := getSessionHandler(c)
	username := session.Values["username"]
	return c.String(http.StatusOK, username.(string))
}

func SessionMiddleware(store *sessions.CookieStore) {

}
