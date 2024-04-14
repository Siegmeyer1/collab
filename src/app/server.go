package app

import (
	"diploma/src/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//type Server struct {
//	e        *echo.Echo
//	sessions *session.Repository
//}
//
//func (s *Server) Start(address string) error {
//	return s.e.Start(address)
//}

func BuildServer() (*echo.Echo, error) {
	e := echo.New()

	e.Use(middleware.Recover())

	h := &Handler{sessions: session.NewRepository()}
	h.SetupHandles(e)

	//s := &Server{e: e, sessions: session.NewRepository()}

	return e, nil
}
