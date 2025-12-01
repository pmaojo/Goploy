package tui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDomainsInput(t *testing.T) {
	input := "example.com, www.example.com\n api.example.com , ,"
	domains := parseDomainsInput(input)

	require.Equal(t, []string{"example.com", "www.example.com", "api.example.com"}, domains)
}
