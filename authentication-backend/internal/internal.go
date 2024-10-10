package internal

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type basicDBUser struct {
	Username string `json:"username" db:"username"`
}

func HandleServeEvent(e *core.ServeEvent) error {
	e.Router.POST("/internal/request-stream", func(c echo.Context) error {
		payload := webrtc.WebhookPayload{}
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			e.App.Logger().Error("Error reading request body: " + err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading request body"})
		}
		json.Unmarshal(body, &payload)
		streamkey := payload.StreamKey
		e.App.Logger().Debug("Requesting stream with key: " + streamkey)

		user := &basicDBUser{}
		err = e.App.Dao().DB().
			Select("username").
			From("users").
			InnerJoin("streamkeys", dbx.NewExp("users.streamkey_id = streamkeys.id")).
			Where(dbx.Like("streamkey", streamkey)).One(user)
		if err != nil {
			e.App.Logger().Error("Error finding user: " + err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error finding user"})
		}

		e.App.Logger().Debug("User found: " + user.Username)
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello " + user.Username})
	} /* optional middlewares */)

	return nil
}
