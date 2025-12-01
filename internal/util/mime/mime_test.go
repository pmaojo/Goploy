package mime_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pmaojo/goploy/internal/util"
	"github.com/pmaojo/goploy/internal/util/mime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnownMIME(t *testing.T) {
	setProjectRootForMimeTests(t)

	filePath := filepath.Join(util.GetProjectRootDir(), "test", "testdata", "example.jpg")

	var detectedMIME mime.MIME
	var err error
	detectedMIME, err = mimetype.DetectFile(filePath)
	require.NoError(t, err)

	var knownMIME mime.MIME = &mime.KnownMIME{
		MimeType:      "image/jpeg",
		FileExtension: ".jpg",
	}

	assert.Equal(t, detectedMIME.Extension(), knownMIME.Extension())
	assert.Equal(t, detectedMIME.String(), knownMIME.String())
	assert.True(t, knownMIME.Is(detectedMIME.String()))
}

func setProjectRootForMimeTests(t *testing.T) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)

	root, err := util.FindProjectRoot(filepath.Dir(file))
	require.NoError(t, err)
	t.Setenv("PROJECT_ROOT_DIR", root)
}
