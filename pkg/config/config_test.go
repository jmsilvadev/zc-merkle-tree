package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	got := New(context.Background(), ":4000", "leveldb", "", nil, &zap.Logger{})
	if got.ServerPort != ":4000" {
		t.Errorf("Got and Expected are not equals. Got: %v, expected: :4000", got.ServerPort)
	}
}

func TestGetDeaultConfig(t *testing.T) {
	config := GetDefaultConfig()
	if config.ServerPort == "" {
		t.Errorf("Got and Expected are not equals. got: '', expected: !''")
	}
}

func TestGetEnv(t *testing.T) {
	v := getEnv("a", "b")
	require.Equal(t, "b", v)
}
