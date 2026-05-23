package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func setLogLevel(level string) error {
	parsed, err := parseLogLevel(level)
	if err != nil {
		return err
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(parsed)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	zlog.Logger = zerolog.New(writer).With().Timestamp().Logger().Level(parsed)
	return nil
}

func parseLogLevel(level string) (zerolog.Level, error) {
	normalized := strings.ToLower(strings.TrimSpace(level))
	switch normalized {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "", "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf(
			"invalid log level %q (allowed: trace|debug|info|warn|error|fatal|panic)",
			level,
		)
	}
}

func loadDotEnv() error {
	dotEnvPath := strings.TrimSpace(os.Getenv("BUGLENS_DOTENV_PATH"))
	if dotEnvPath == "" {
		dotEnvPath = ".env"
	}

	_, err := os.Stat(dotEnvPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return godotenv.Load(dotEnvPath)
}

func getEnvDefault(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return fallback
}

func getEnvDefaultInt(k string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
