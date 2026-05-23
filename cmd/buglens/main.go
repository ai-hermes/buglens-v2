package main

import (
	"fmt"
	"os"
)

func main() {
	if err := loadDotEnv(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to load .env:", err)
		os.Exit(1)
	}
	if err := setLogLevel(getEnvDefault("BUGLENS_LOG_LEVEL", "INFO")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
