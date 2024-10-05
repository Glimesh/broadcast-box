package internal

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type User struct {
	Username string `json:"username" db:"username"`
}

func HandleServeEvent(e *core.ServeEvent) error {
	e.Router.GET("/internal/request-stream/:streamkey", func(c echo.Context) error {
		streamkey := c.PathParam("streamkey")

		e.App.Logger().Debug("Requesting stream with key: " + streamkey)

		user := &User{}
		err := e.App.Dao().DB().Select("username").From("users").Where(dbx.NewExp("streamkey.streamkey = " + streamkey)).One(user)
		if err != nil {
			e.App.Logger().Error("Error finding user: " + err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error finding user"})
		}

		e.App.Logger().Debug("User found: " + user.Username)
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello " + user.Username})
	} /* optional middlewares */)

	return nil
}
