package main

import (
	"fmt"
	"strings"

	server "github.com/jmsilvadev/zc/cmd/server/internal"
	"github.com/jmsilvadev/zc/pkg/config"
	"github.com/jmsilvadev/zc/pkg/db"
	"github.com/jmsilvadev/zc/pkg/leveldb"
	"github.com/jmsilvadev/zc/pkg/scylladb"
)

const defaultScyllaKS = "zc" // TODO: put this as env var in config

func main() {
	c := config.GetDefaultConfig()
	run(c)
}

func run(c *config.Config) error {
	if strings.ToLower(c.DbEngine) != "leveldb" && strings.ToLower(c.DbEngine) != "scylladb" {
		return fmt.Errorf("invalid db engine, valid values: leveldb, scylladb")
	}

	var db db.Database
	var err error

	db, err = leveldb.New(c.DbPath)
	if strings.ToLower(c.DbEngine) == "scylladb" {
		db, err = scylladb.New(c.ScyllaHosts, defaultScyllaKS)
	}

	if err != nil {
		return fmt.Errorf("DB ERROR: %s", err.Error())
	}

	s := server.NewServer(c, db)
	s.Start()
	return nil
}
