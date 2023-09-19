package router

import (
	"net/http"

	"github.com/NikhilSharmaWe/quasar/model"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (app *Application) Router() *echo.Echo {
	e := echo.New()

	// this will remove the trailing slash from the req url
	// to match the path specified in the router without it
	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(app.createSessionMiddleware)

	e.Static("/assets", "./public")

	e.GET("/", ServeHTML("./public/login/index.html"), app.ifAlreadyLogined)
	e.GET("/signup", ServeHTML("./public/signup/index.html"), app.ifAlreadyLogined)
	e.GET("/meets", ServeHTML("./public/meets/index.html"), app.ifNotLogined)
	e.GET("/logout", app.HandleLogout)

	e.POST("/", app.HandleSignIn)
	e.POST("/signup", app.HandleSignUp)

	return e
}

func ServeHTML(htmlPath string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.File(htmlPath)
	}
}

func (app *Application) HandleSignUp(c echo.Context) error {
	user, err := userFromForm(c)
	if err != nil {
		return err
	}

	if err := app.UserRepo.SaveAccount(user); err != nil {
		if err.Error() == model.AlreadyExistsErr {
			return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
		}

		app.Logger.Println("error:", err)
		return err
	}

	if err := setSession(c); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/meets/")
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
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/meets/")
}

func (app *Application) HandleLogout(c echo.Context) error {
	if err := clearSessionHandler(c); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
