package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestDotEnvOverride(t *testing.T) {
	wd, _ := os.Getwd()
	// internal/config -> root is ../..
	os.Setenv("PROJECT_ROOT_DIR", filepath.Join(wd, "../.."))
	// The util.GetProjectRootDir uses sync.Once, so we can't easily reset it if it was already called.
	// However, go test runs each package in a separate process (usually).
	// But if parallel tests run, it might be tricky.
	// Let's hope sync.Once hasn't triggered yet.

	assert.Empty(t, os.Getenv("IS_THIS_A_TEST_ENV"))

	orgPsqlUser := os.Getenv("PSQL_USER")

	config.DotEnvTryLoad(
		filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env1.local"),
		func(k string, v string) error { t.Setenv(k, v); return nil })

	assert.Equal(t, "yes", os.Getenv("IS_THIS_A_TEST_ENV"))
	assert.Equal(t, "dotenv_override_psql_user", os.Getenv("PSQL_USER"))
	assert.Equal(t, orgPsqlUser, os.Getenv("ORIGINAL_PSQL_USER"))

	// override works as expected?
	config.DotEnvTryLoad(
		filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env2.local"),
		func(k string, v string) error { t.Setenv(k, v); return nil })

	assert.Equal(t, "yes still", os.Getenv("IS_THIS_A_TEST_ENV"))
	assert.NotEqual(t, "dotenv_override_psql_user", os.Getenv("PSQL_USER"))
	assert.Equal(t, orgPsqlUser, os.Getenv("PSQL_USER"), "Reset to original does not work!")
}

func TestNoopEnvNotFound(t *testing.T) {
	wd, _ := os.Getwd()
	os.Setenv("PROJECT_ROOT_DIR", filepath.Join(wd, "../.."))

	assert.NotPanics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.does.not.exist"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does not panic on file inexistance")
}

func TestEmptyEnv(t *testing.T) {
	wd, _ := os.Getwd()
	os.Setenv("PROJECT_ROOT_DIR", filepath.Join(wd, "../.."))

	assert.NotPanics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.local.sample"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does not panic on file inexistance")

	assert.Empty(t, os.Getenv("EMPTY_VARIABLE_INIT"), "should be empty")
}

func TestPanicsOnEnvMalform(t *testing.T) {
	wd, _ := os.Getwd()
	os.Setenv("PROJECT_ROOT_DIR", filepath.Join(wd, "../.."))

	assert.Panics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.local.malformed"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does panic on file malform")
}
