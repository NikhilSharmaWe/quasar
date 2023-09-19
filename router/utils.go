package router

import (
	"context"
	"log"
	"os"

	"github.com/NikhilSharmaWe/quasar/model"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	Context     context.Context
	Logger      *log.Logger
	UserRepo    *model.GenericRepo[model.User]
	CookieStore *sessions.CookieStore
}

func NewApplication() *Application {
	var (
		ctx      = context.Background()
		logger   = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		userRepo = model.GenericRepo[model.User]{
			Collection: "user",
		}
		cookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET_KEY")))
	)

	application := &Application{
		Context:     ctx,
		Logger:      logger,
		UserRepo:    &userRepo,
		CookieStore: cookieStore,
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

func userFromForm(c echo.Context) (*model.User, error) {
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
