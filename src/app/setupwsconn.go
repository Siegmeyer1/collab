package app

import (
	"collab/src/logging"
	"collab/src/session"
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"nhooyr.io/websocket"
	"time"
)

func (h *Handler) setupWSConn(c echo.Context) error {
	roomName, err := h.getRoomName(c)
	if err != nil {
		logging.Exception(err)
		return err
	}

	opts := &websocket.AcceptOptions{OriginPatterns: []string{"localhost:*"}}

	sess, err := h.sessions.GetOrCreateSession(roomName)
	if err != nil {
		return err
	}

	wsClient := session.NewClient(roomName, sess)

	conn, err := websocket.Accept(c.Response().Writer, c.Request(), opts)
	if err != nil {
		logging.Error("room name: %s, err: %v", roomName, err)
		return err
	}

	err = wsClient.Start(context.Background(), conn)
	defer wsClient.Close()

	// just error handling beyond this point
	var closeErr websocket.CloseError
	if errors.As(err, &closeErr) {
		closeStatus := websocket.CloseStatus(closeErr)
		if closeStatus == websocket.StatusNormalClosure ||
			closeStatus == websocket.StatusGoingAway ||
			closeStatus == websocket.StatusProtocolError {
			logging.Info("websocket normal close. Code: %d\n", closeStatus)
		} else {
			now := time.Now().String()
			logging.Info("[%s] close Reason: %s, Code: %d\n", now, closeErr.Reason, closeErr.Code)
		}
		err = conn.Close(closeErr.Code, closeErr.Reason)
		if err != nil {
			logging.Error("closing ws conn: %v\n", err)
		}
		return nil
	}

	if err != nil {
		logging.Exception(err)
	}

	return nil
}
