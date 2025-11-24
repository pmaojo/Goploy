package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"allaboutapps.dev/aw/go-starter/internal/deployment"
	"allaboutapps.dev/aw/go-starter/internal/mailer"
	"allaboutapps.dev/aw/go-starter/internal/util"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type Router struct {
	Routes        []*echo.Route
	Root          *echo.Group
	Management    *echo.Group
	APIV1Projects *echo.Group
	WellKnown     *echo.Group
}

// Server is a central struct keeping all the dependencies.
type Server struct {
	// skip wire:
	// -> initialized with router.Init(s) function
	Echo   *echo.Echo `wire:"-"`
	Router *Router    `wire:"-"`

	Config         config.Server
	GoployConfig   *config.GoployConfig
	Deployment     deployment.Controller
	Mailer         *mailer.Mailer
}

func NewServer(config config.Server, goployConfig *config.GoployConfig, mailer *mailer.Mailer, dep deployment.Controller) *Server {
	s := &Server{
		Config:       config,
		GoployConfig: goployConfig,
		Mailer:       mailer,
		Deployment:   dep,
	}

	return s
}

func (s *Server) Ready() bool {
	if err := util.IsStructInitialized(s); err != nil {
		log.Debug().Err(err).Msg("Server is not fully initialized")
		return false
	}

	return true
}

func (s *Server) Start() error {
	if !s.Ready() {
		return errors.New("server is not ready")
	}

	if err := s.Echo.Start(s.Config.Echo.ListenAddress); err != nil {
		return fmt.Errorf("failed to start echo server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) []error {
	log.Warn().Msg("Shutting down server")

	var errs []error

	if s.Echo != nil {
		log.Debug().Msg("Shutting down echo server")

		if err := s.Echo.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Failed to shutdown echo server")
			errs = append(errs, err)
		}
	}

	return errs
}
