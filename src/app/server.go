package app

import (
	"collab/src/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func BuildServer() (*echo.Echo, error) {
	e := echo.New()

	e.Use(middleware.Recover())

	h := &Handler{sessions: session.NewRepository()}
	h.SetupHandles(e)

	return e, nil
}
