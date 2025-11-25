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
	// Ensure clean state
	t.Setenv("IS_THIS_A_TEST_ENV", "")
	assert.Empty(t, os.Getenv("IS_THIS_A_TEST_ENV"))

	// Since PSQL_USER might not be set in this environment, handle it accordingly.
	// We want to test that .env overrides existing env vars, or sets new ones.

	// Set a baseline
	t.Setenv("PSQL_USER", "original_user")
	orgPsqlUser := "original_user"

	config.DotEnvTryLoad(
		filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env1.local"),
		func(k string, v string) error { t.Setenv(k, v); return nil })

	assert.Equal(t, "yes", os.Getenv("IS_THIS_A_TEST_ENV"))
	assert.Equal(t, "dotenv_override_psql_user", os.Getenv("PSQL_USER"))

	// The .env1.local file sets ORIGINAL_PSQL_USER=${PSQL_USER}.
	// The expansion happens based on the *current* env when loading.
	// If PSQL_USER was "original_user" before load, then ORIGINAL_PSQL_USER should be "original_user".
	assert.Equal(t, orgPsqlUser, os.Getenv("ORIGINAL_PSQL_USER"))

	// override works as expected?
	config.DotEnvTryLoad(
		filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env2.local"),
		func(k string, v string) error { t.Setenv(k, v); return nil })

	assert.Equal(t, "yes still", os.Getenv("IS_THIS_A_TEST_ENV"))

	// .env2.local sets PSQL_USER=${ORIGINAL_PSQL_USER}
	// Since ORIGINAL_PSQL_USER is "original_user" (set in env1 load), PSQL_USER should revert to "original_user".
	assert.Equal(t, orgPsqlUser, os.Getenv("PSQL_USER"), "Reset to original does not work!")
}

func TestNoopEnvNotFound(t *testing.T) {
	assert.NotPanics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.does.not.exist"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does not panic on file inexistance")
}

func TestEmptyEnv(t *testing.T) {
	assert.NotPanics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.local.sample"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does not panic on file inexistance")

	assert.Empty(t, os.Getenv("EMPTY_VARIABLE_INIT"), "should be empty")
}

func TestPanicsOnEnvMalform(t *testing.T) {
	assert.Panics(t, assert.PanicTestFunc(func() {
		config.DotEnvTryLoad(
			filepath.Join(util.GetProjectRootDir(), "/internal/config/testdata/.env.local.malformed"),
			func(k string, v string) error { t.Setenv(k, v); return nil },
		)
	}), "does panic on file malform")
}
