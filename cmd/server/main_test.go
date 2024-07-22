package main

import (
	"testing"

	"github.com/jmsilvadev/zc/pkg/config"
)

func TestRun(t *testing.T) {
	c := config.GetDefaultConfig()
	c.DbEngine = "non-evistent"
	run(c)

	c.DbEngine = "scylladb"
	run(c)
}
