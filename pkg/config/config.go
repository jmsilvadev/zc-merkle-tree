package config

import (
	"context"
	"os"
	"strings"

	"github.com/jmsilvadev/zc/pkg/logger"
)

// Default Values
var (
	dbEngine    = "leveldb"
	dbPath      = "/tmp/sc.db"
	serverPort  = ":5000"
	loggerLevel = "INFO"
	scyllaHosts = "localhost"
)

type Config struct {
	DbPath      string
	DbEngine    string
	ServerPort  string
	ScyllaHosts []string
	Logger      logger.Logger
}

func New(ctx context.Context, port, dbEngine, dbPath string, scyllaHosts []string, logger logger.Logger) *Config {
	return &Config{
		ServerPort:  port,
		Logger:      logger,
		DbPath:      dbPath,
		DbEngine:    dbEngine,
		ScyllaHosts: scyllaHosts,
	}
}

func GetDefaultConfig() *Config {
	serverPort = getEnv("SERVER_PORT", serverPort)
	loggerLevel = getEnv("LOG_LEVEL", loggerLevel)
	dbPath = getEnv("DB_PATH", dbPath)
	dbEngine = getEnv("DB_ENGINE", dbEngine)
	scyllaHosts = getEnv("SCYLLA_HOSTS", scyllaHosts)

	level := logger.LEVEL_ERROR
	if loggerLevel == "INFO" {
		level = logger.LEVEL_INFO
	}
	if loggerLevel == "WARN" {
		level = logger.LEVEL_WARN
	}
	if loggerLevel == "DEBUG" {
		level = logger.LEVEL_DEBUG
	}

	ctx := context.Background()
	log := logger.New(level)

	hosts := strings.Split(scyllaHosts, ",")

	config := New(ctx, serverPort, dbEngine, dbPath, hosts, log)

	return config
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
