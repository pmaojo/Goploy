package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/goploy/internal/mailer/transport"
	"github.com/pmaojo/goploy/internal/util"
	"github.com/rs/zerolog"
)

// EchoServer configures the Echo web framework.
type EchoServer struct {
	Debug                          bool
	ListenAddress                  string
	HideInternalServerErrorDetails bool
	BaseURL                        string
	EnableCORSMiddleware           bool
	EnableLoggerMiddleware         bool
	EnableRecoverMiddleware        bool
	EnableRequestIDMiddleware      bool
	EnableTrailingSlashMiddleware  bool
	EnableSecureMiddleware         bool
	EnableCacheControlMiddleware   bool
	SecureMiddleware               EchoServerSecureMiddleware
	WebTemplatesViewsBaseDirAbs    string
}

// PprofServer configures the pprof debugging server.
type PprofServer struct {
	Enable                      bool
	EnableManagementKeyAuth     bool
	RuntimeBlockProfileRate     int
	RuntimeMutexProfileFraction int
}

// EchoServerSecureMiddleware represents a subset of echo's secure middleware config relevant to the app server.
// https://github.com/labstack/echo/blob/master/middleware/secure.go
type EchoServerSecureMiddleware struct {
	XSSProtection         string
	ContentTypeNosniff    string
	XFrameOptions         string
	HSTSMaxAge            int
	HSTSExcludeSubdomains bool
	ContentSecurityPolicy string
	CSPReportOnly         bool
	HSTSPreloadEnabled    bool
	ReferrerPolicy        string
}

// ManagementServer configures the internal management API.
type ManagementServer struct {
	Secret                  string `json:"-"` // sensitive
	EnableMetrics           bool
}

// LoggerServer configures the application logging.
type LoggerServer struct {
	Level              zerolog.Level
	RequestLevel       zerolog.Level
	LogRequestBody     bool
	LogRequestHeader   bool
	LogRequestQuery    bool
	LogResponseBody    bool
	LogResponseHeader  bool
	LogCaller          bool
	PrettyPrintConsole bool
}

// Server holds the complete configuration for the application server, including
// HTTP, logging, mailer, and Goploy-specific settings.
type Server struct {
	Echo       EchoServer
	Pprof      PprofServer
	Management ManagementServer
	Mailer     Mailer
	SMTP       transport.SMTPMailTransportConfig
	Logger     LoggerServer
	Goploy     GoployServer
}

// GoployServer holds configuration specific to Goploy functionality.
type GoployServer struct {
	APIKey string `json:"-"` // sensitive
}

// DefaultServiceConfigFromEnv returns the server config as parsed from environment variables
// and their respective defaults defined below.
// We don't expect that ENV_VARs change while we are running our application or our tests
// (and it would be a bad thing to do anyways with parallel testing).
// Do NOT use os.Setenv / os.Unsetenv in tests utilizing DefaultServiceConfigFromEnv()!
func DefaultServiceConfigFromEnv() Server {
	// An `.env.local` file in your project root can override the currently set ENV variables.
	//
	// We never automatically apply `.env.local` when running "go test" as these ENV variables
	// may be sensitive (e.g. secrets to external APIs) and applying them modifies the process
	// global "os.Env" state (it should be applied via t.SetEnv instead).
	//
	// If you need dotenv ENV variables available in a test, do that explicitly within that
	// test before executing DefaultServiceConfigFromEnv (or test.WithTestServer).
	// See /internal/test/helper_dot_env.go: test.DotEnvLoadLocalOrSkipTest(t)
	if !testing.Testing() {
		DotEnvTryLoad(filepath.Join(util.GetProjectRootDir(), ".env.local"), os.Setenv)
	}

	return Server{
		Echo: EchoServer{
			Debug:                          util.GetEnvAsBool("SERVER_ECHO_DEBUG", false),
			ListenAddress:                  util.GetEnv("SERVER_ECHO_LISTEN_ADDRESS", ":8080"),
			HideInternalServerErrorDetails: util.GetEnvAsBool("SERVER_ECHO_HIDE_INTERNAL_SERVER_ERROR_DETAILS", true),
			BaseURL:                        util.GetEnv("SERVER_ECHO_BASE_URL", "http://localhost:8080"),
			EnableCORSMiddleware:           util.GetEnvAsBool("SERVER_ECHO_ENABLE_CORS_MIDDLEWARE", true),
			EnableLoggerMiddleware:         util.GetEnvAsBool("SERVER_ECHO_ENABLE_LOGGER_MIDDLEWARE", true),
			EnableRecoverMiddleware:        util.GetEnvAsBool("SERVER_ECHO_ENABLE_RECOVER_MIDDLEWARE", true),
			EnableRequestIDMiddleware:      util.GetEnvAsBool("SERVER_ECHO_ENABLE_REQUEST_ID_MIDDLEWARE", true),
			EnableTrailingSlashMiddleware:  util.GetEnvAsBool("SERVER_ECHO_ENABLE_TRAILING_SLASH_MIDDLEWARE", true),
			EnableSecureMiddleware:         util.GetEnvAsBool("SERVER_ECHO_ENABLE_SECURE_MIDDLEWARE", true),
			EnableCacheControlMiddleware:   util.GetEnvAsBool("SERVER_ECHO_ENABLE_CACHE_CONTROL_MIDDLEWARE", true),
			// see https://echo.labstack.com/middleware/secure
			// see https://github.com/labstack/echo/blob/master/middleware/secure.go
			SecureMiddleware: EchoServerSecureMiddleware{
				XSSProtection:         util.GetEnv("SERVER_ECHO_SECURE_MIDDLEWARE_XSS_PROTECTION", "1; mode=block"),
				ContentTypeNosniff:    util.GetEnv("SERVER_ECHO_SECURE_MIDDLEWARE_CONTENT_TYPE_NOSNIFF", "nosniff"),
				XFrameOptions:         util.GetEnv("SERVER_ECHO_SECURE_MIDDLEWARE_X_FRAME_OPTIONS", "SAMEORIGIN"),
				HSTSMaxAge:            util.GetEnvAsInt("SERVER_ECHO_SECURE_MIDDLEWARE_HSTS_MAX_AGE", 0),
				HSTSExcludeSubdomains: util.GetEnvAsBool("SERVER_ECHO_SECURE_MIDDLEWARE_HSTS_EXCLUDE_SUBDOMAINS", false),
				ContentSecurityPolicy: util.GetEnv("SERVER_ECHO_SECURE_MIDDLEWARE_CONTENT_SECURITY_POLICY", ""),
				CSPReportOnly:         util.GetEnvAsBool("SERVER_ECHO_SECURE_MIDDLEWARE_CSP_REPORT_ONLY", false),
				HSTSPreloadEnabled:    util.GetEnvAsBool("SERVER_ECHO_SECURE_MIDDLEWARE_HSTS_PRELOAD_ENABLED", false),
				ReferrerPolicy:        util.GetEnv("SERVER_ECHO_SECURE_MIDDLEWARE_REFERRER_POLICY", ""),
			},
			WebTemplatesViewsBaseDirAbs: util.GetEnv("SERVER_ECHO_WEB_TEMPLATES_VIEWS_BASE_DIR_ABS", filepath.Join(util.GetProjectRootDir(), "/web/templates/views")),
		},
		Pprof: PprofServer{
			// https://golang.org/pkg/net/http/pprof/
			Enable:                      util.GetEnvAsBool("SERVER_PPROF_ENABLE", false),
			EnableManagementKeyAuth:     util.GetEnvAsBool("SERVER_PPROF_ENABLE_MANAGEMENT_KEY_AUTH", true),
			RuntimeBlockProfileRate:     util.GetEnvAsInt("SERVER_PPROF_RUNTIME_BLOCK_PROFILE_RATE", 0),
			RuntimeMutexProfileFraction: util.GetEnvAsInt("SERVER_PPROF_RUNTIME_MUTEX_PROFILE_FRACTION", 0),
		},
		Management: ManagementServer{
			Secret:        util.GetMgmtSecret("SERVER_MANAGEMENT_SECRET"),
			EnableMetrics: util.GetEnvAsBool("SERVER_MANAGEMENT_ENABLE_METRICS", false),
		},
		Mailer: Mailer{
			DefaultSender:               util.GetEnv("SERVER_MAILER_DEFAULT_SENDER", "go-starter@example.com"),
			Send:                        util.GetEnvAsBool("SERVER_MAILER_SEND", true),
			WebTemplatesEmailBaseDirAbs: util.GetEnv("SERVER_MAILER_WEB_TEMPLATES_EMAIL_BASE_DIR_ABS", filepath.Join(util.GetProjectRootDir(), "/web/templates/email")), // /app/web/templates/email
			Transporter:                 util.GetEnvEnum("SERVER_MAILER_TRANSPORTER", MailerTransporterMock.String(), []string{MailerTransporterSMTP.String(), MailerTransporterMock.String()}),
		},
		SMTP: transport.SMTPMailTransportConfig{
			Host:       util.GetEnv("SERVER_SMTP_HOST", "mailhog"),
			Port:       util.GetEnvAsInt("SERVER_SMTP_PORT", 1025),
			Username:   util.GetEnv("SERVER_SMTP_USERNAME", ""),
			Password:   util.GetEnv("SERVER_SMTP_PASSWORD", ""),
			AuthType:   transport.SMTPAuthTypeFromString(util.GetEnv("SERVER_SMTP_AUTH_TYPE", transport.SMTPAuthTypeNone.String())),
			Encryption: transport.SMTPEncryption(util.GetEnvEnum("SERVER_SMTP_ENCRYPTION", transport.SMTPEncryptionNone.String(), []string{transport.SMTPEncryptionNone.String(), transport.SMTPEncryptionTLS.String(), transport.SMTPEncryptionStartTLS.String()})),
			TLSConfig:  nil,
		},
		Logger: LoggerServer{
			Level:              util.LogLevelFromString(util.GetEnv("SERVER_LOGGER_LEVEL", zerolog.DebugLevel.String())),
			RequestLevel:       util.LogLevelFromString(util.GetEnv("SERVER_LOGGER_REQUEST_LEVEL", zerolog.DebugLevel.String())),
			LogRequestBody:     util.GetEnvAsBool("SERVER_LOGGER_LOG_REQUEST_BODY", false),
			LogRequestHeader:   util.GetEnvAsBool("SERVER_LOGGER_LOG_REQUEST_HEADER", false),
			LogRequestQuery:    util.GetEnvAsBool("SERVER_LOGGER_LOG_REQUEST_QUERY", false),
			LogResponseBody:    util.GetEnvAsBool("SERVER_LOGGER_LOG_RESPONSE_BODY", false),
			LogResponseHeader:  util.GetEnvAsBool("SERVER_LOGGER_LOG_RESPONSE_HEADER", false),
			LogCaller:          util.GetEnvAsBool("SERVER_LOGGER_LOG_CALLER", false),
			PrettyPrintConsole: util.GetEnvAsBool("SERVER_LOGGER_PRETTY_PRINT_CONSOLE", false),
		},
		Goploy: GoployServer{
			APIKey: util.GetEnv("GOPLOY_API_KEY", ""),
		},
	}
}
