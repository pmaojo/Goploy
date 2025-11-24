package router

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"allaboutapps.dev/aw/go-starter/internal/api"
	"allaboutapps.dev/aw/go-starter/internal/api/handlers"
	"allaboutapps.dev/aw/go-starter/internal/api/middleware"
	"allaboutapps.dev/aw/go-starter/internal/api/router/templates"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	// #nosec G108 - pprof handlers (conditionally made available via http.DefaultServeMux)
	"net/http/pprof"
)

func Init(s *api.Server) error {
	s.Echo = echo.New()

	viewsRenderer := &echoRenderer{
		templates: map[templates.ViewTemplate]*template.Template{},
	}

	files, err := os.ReadDir(s.Config.Echo.WebTemplatesViewsBaseDirAbs)
	if err != nil {
		// Log but don't fail, maybe we don't need views
		log.Warn().Err(err).Msg("Failed to read views templates dir, skipping")
	} else {
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			templateName := file.Name()
			t, err := template.New(templateName).ParseGlob(filepath.Join(s.Config.Echo.WebTemplatesViewsBaseDirAbs, templateName))
			if err != nil {
				return fmt.Errorf("failed to parse template file: %w", err)
			}

			viewsRenderer.templates[templates.ViewTemplate(templateName)] = t
		}
	}

	s.Echo.Renderer = viewsRenderer

	s.Echo.Debug = s.Config.Echo.Debug
	s.Echo.HideBanner = true
	s.Echo.Logger.SetOutput(&echoLogger{level: s.Config.Logger.RequestLevel, log: log.With().Str("component", "echo").Logger()})

	// Use default NotFoundHandler or simple one as config.Frontend was removed
	echo.NotFoundHandler = func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Not Found"})
	}

	// Default HTTP error handler
	s.Echo.HTTPErrorHandler = func(err error, c echo.Context) {
		s.Echo.DefaultHTTPErrorHandler(err, c)
	}

	// ---
	// General middleware
	if s.Config.Management.EnableMetrics {
		s.Echo.Use(echoprometheus.NewMiddleware(""))
	}

	if s.Config.Echo.EnableTrailingSlashMiddleware {
		s.Echo.Pre(echoMiddleware.RemoveTrailingSlash())
	}

	if s.Config.Echo.EnableRecoverMiddleware {
		s.Echo.Use(echoMiddleware.RecoverWithConfig(echoMiddleware.RecoverConfig{
			LogErrorFunc: middleware.LogErrorFuncWithRequestInfo,
		}))
	}

	if s.Config.Echo.EnableSecureMiddleware {
		s.Echo.Use(echoMiddleware.SecureWithConfig(echoMiddleware.SecureConfig{
			Skipper:               echoMiddleware.DefaultSecureConfig.Skipper,
			XSSProtection:         s.Config.Echo.SecureMiddleware.XSSProtection,
			ContentTypeNosniff:    s.Config.Echo.SecureMiddleware.ContentTypeNosniff,
			XFrameOptions:         s.Config.Echo.SecureMiddleware.XFrameOptions,
			HSTSMaxAge:            s.Config.Echo.SecureMiddleware.HSTSMaxAge,
			HSTSExcludeSubdomains: s.Config.Echo.SecureMiddleware.HSTSExcludeSubdomains,
			ContentSecurityPolicy: s.Config.Echo.SecureMiddleware.ContentSecurityPolicy,
			CSPReportOnly:         s.Config.Echo.SecureMiddleware.CSPReportOnly,
			HSTSPreloadEnabled:    s.Config.Echo.SecureMiddleware.HSTSPreloadEnabled,
			ReferrerPolicy:        s.Config.Echo.SecureMiddleware.ReferrerPolicy,
		}))
	}

	if s.Config.Echo.EnableRequestIDMiddleware {
		s.Echo.Use(echoMiddleware.RequestID())
	}

	if s.Config.Echo.EnableLoggerMiddleware {
		s.Echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Level:             s.Config.Logger.RequestLevel,
			LogRequestBody:    s.Config.Logger.LogRequestBody,
			LogRequestHeader:  s.Config.Logger.LogRequestHeader,
			LogRequestQuery:   s.Config.Logger.LogRequestQuery,
			LogResponseBody:   s.Config.Logger.LogResponseBody,
			LogResponseHeader: s.Config.Logger.LogResponseHeader,
			LogCaller:         s.Config.Logger.LogCaller,
			RequestBodyLogSkipper: func(req *http.Request) bool {
				return middleware.DefaultRequestBodyLogSkipper(req)
			},
			ResponseBodyLogSkipper: func(req *http.Request, res *echo.Response) bool {
				return middleware.DefaultResponseBodyLogSkipper(req, res)
			},
			Skipper: func(c echo.Context) bool {
				// We skip logging of readiness and liveness endpoints
				switch c.Path() {
				case "/-/ready", "/-/healthy":
					return true
				}
				return false
			},
		}))
	}

	if s.Config.Echo.EnableCORSMiddleware {
		s.Echo.Use(echoMiddleware.CORS())
	}

	if s.Config.Echo.EnableCacheControlMiddleware {
		s.Echo.Use(middleware.CacheControl())
	}

	if s.Config.Pprof.Enable {
		pprofAuthMiddleware := middleware.Noop()

		if s.Config.Pprof.EnableManagementKeyAuth {
			pprofAuthMiddleware = echoMiddleware.KeyAuthWithConfig(echoMiddleware.KeyAuthConfig{
				KeyLookup: "query:mgmt-secret",
				Validator: func(key string, _ echo.Context) (bool, error) {
					return key == s.Config.Management.Secret, nil
				},
			})
		}

		s.Echo.GET("/debug/pprof", echo.WrapHandler(http.HandlerFunc(pprof.Index)), pprofAuthMiddleware)
		s.Echo.Any("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux), pprofAuthMiddleware)

		log.Warn().Bool("EnableManagementKeyAuth", s.Config.Pprof.EnableManagementKeyAuth).Msg("Pprof http handlers are available at /debug/pprof")

		if s.Config.Pprof.RuntimeBlockProfileRate != 0 {
			runtime.SetBlockProfileRate(s.Config.Pprof.RuntimeBlockProfileRate)
			log.Warn().Int("RuntimeBlockProfileRate", s.Config.Pprof.RuntimeBlockProfileRate).Msg("Pprof runtime.SetBlockProfileRate")
		}

		if s.Config.Pprof.RuntimeMutexProfileFraction != 0 {
			runtime.SetMutexProfileFraction(s.Config.Pprof.RuntimeMutexProfileFraction)
			log.Warn().Int("RuntimeMutexProfileFraction", s.Config.Pprof.RuntimeMutexProfileFraction).Msg("Pprof runtime.SetMutexProfileFraction")
		}
	}

	// ---
	// Initialize our general groups and set middleware to use above them
	s.Router = &api.Router{
		Routes: nil, // will be populated by handlers.AttachAllRoutes(s)

		// Unsecured base group available at /**
		Root: s.Echo.Group(""),

		// Management endpoints, uncacheable, secured by key auth (query param), available at /-/**
		Management: s.Echo.Group("/-", echoMiddleware.KeyAuthWithConfig(echoMiddleware.KeyAuthConfig{
			KeyLookup: "query:mgmt-secret",
			Validator: func(key string, _ echo.Context) (bool, error) {
				return key == s.Config.Management.Secret, nil
			},
			Skipper: func(c echo.Context) bool {
				//nolint:gocritic
				switch c.Path() {
				case "/-/ready":
					return true
				}
				return false
			},
		}), middleware.NoCache()),

		// Goploy Project Endpoints
		// Secured by Bearer token (GOPLOY_API_KEY)
		APIV1Projects: s.Echo.Group("/api/v1/projects", echoMiddleware.KeyAuthWithConfig(echoMiddleware.KeyAuthConfig{
			KeyLookup: "header:Authorization",
			AuthScheme: "Bearer",
			Validator: func(key string, c echo.Context) (bool, error) {
				return key == s.Config.Goploy.APIKey, nil
			},
		})),

		WellKnown: s.Echo.Group("/.well-known"),
	}

	// ---
	// Finally attach our handlers
	handlers.AttachAllRoutes(s)

	if s.Config.Management.EnableMetrics {
		log.Info().Msg("Metrics enabled and available under /metrics")
		s.Echo.GET("/metrics", echoprometheus.NewHandler())
	}

	return nil
}
