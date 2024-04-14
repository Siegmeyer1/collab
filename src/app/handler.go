package app

import (
	"diploma/src/session"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	sessions *session.Repository
}

func (h *Handler) SetupHandles(e *echo.Echo) {
	e.GET("/:room_name", h.setupWSConn)
}

func (h *Handler) getRoomName(c echo.Context) (string, error) {
	roomName := c.Param("room_name")

	// TODO: maybe validate?
	return roomName, nil
}
