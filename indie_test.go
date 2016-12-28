package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestConfigFromFile(t *testing.T) {
	require := require.New(t)

	// When
	config := configFromFile()

	// Then
	require.Equal("golang", config.Image)
}
