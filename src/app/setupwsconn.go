package app

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
	"time"
)

func (h *Handler) setupWSConn(c echo.Context) error {
	roomName, err := h.getRoomName(c)
	if err != nil {
		c.Logger().Error(err)
	}

	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:8080"},
	}

	conn, err := websocket.Accept(c.Response().Writer, c.Request(), opts)
	if err != nil {
		fmt.Printf("room name: %s, err: %v", roomName, err)
		c.Logger().Error(err)
	}

	wsClient := NewClient(roomName, h.sessions, conn)

	var closeErr websocket.CloseError
	err = wsClient.Start(c.Request().Context())

	if errors.As(err, &closeErr) {
		closeStatus := websocket.CloseStatus(closeErr)
		if closeStatus == websocket.StatusNormalClosure ||
			closeStatus == websocket.StatusGoingAway ||
			closeStatus == websocket.StatusProtocolError {
			fmt.Printf("websocket normal close. Code: %d\n", closeStatus)
		} else {
			now := time.Now().String()
			fmt.Printf("[%s] close Reason: %s, Code: %d\n", now, closeErr.Reason, closeErr.Code)
		}
		err = conn.Close(closeErr.Code, closeErr.Reason)
		if err != nil {
			fmt.Printf("closing ws conn: %v\n", err)
		}
		return nil
	}

	if err != nil {
		c.Logger().Error(err)
	}

	return nil
}
