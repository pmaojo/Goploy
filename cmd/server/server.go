package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pmaojo/goploy/internal/api"
	"github.com/pmaojo/goploy/internal/api/router"
	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
	"github.com/pmaojo/goploy/internal/mailer"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Flags struct {
	// Removed DB flags
}

func New() *cobra.Command {
	var flags Flags

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Long: `Starts the RESTful JSON server
	
	Requires configuration through ENV and goploy.yaml.`,
		Run: func(_ *cobra.Command, _ []string) {
			runServer(flags)
		},
	}

	return cmd
}

func runServer(flags Flags) {
	ctx := context.Background()
	ctx = log.With().Str("cmdExecutionId", uuid.New().String()).Logger().WithContext(ctx)

	cfg := config.DefaultServiceConfigFromEnv()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(cfg.Logger.Level)
	if cfg.Logger.PrettyPrintConsole {
		log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = "15:04:05"
		}))
	}

	// Load goploy.yaml
	goployCfg, err := config.LoadGoployConfig("goploy.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load goploy.yaml. Please ensure it exists in the current directory.")
	}

	// Initialize Mailer
	mail, err := mailer.NewWithConfig(cfg.Mailer, cfg.SMTP)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize mailer")
	}

	// Initialize Deployment Controller
	deployer := deployment.NewSSHClient(mail)

	// Initialize Server
	s := api.NewServer(cfg, goployCfg, mail, deployer)

	err = router.Init(s)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize router")
	}

	go func() {
		if err := s.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info().Msg("Server closed")
			} else {
				log.Fatal().Err(err).Msg("Failed to start server")
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if errs := s.Shutdown(shutdownCtx); len(errs) > 0 {
		log.Error().Errs("shutdownErrors", errs).Msg("Failed to gracefully shut down server")
	}
}
