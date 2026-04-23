package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/modules/system/codegen"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "version" {
		cli := codegen.NewCLI(nil, os.Stdout, os.Stderr)
		os.Exit(cli.Run(args))
	}

	configPath := codegen.DetectConfigPath(args)
	format := codegen.DetectFormat(args)

	runtime, err := bootstrap.NewRuntime(configPath)
	if err != nil {
		exitWithError(format, "new runtime", err)
	}

	if sqlDB, err := runtime.DB.DB(); err == nil {
		defer sqlDB.Close()
	}

	cli := codegen.NewCLI(codegen.NewRunner(runtime), os.Stdout, os.Stderr)
	os.Exit(cli.Run(args))
}

func exitWithError(format string, action string, err error) {
	if format == "json" {
		payload, _ := json.MarshalIndent(map[string]any{
			"error": action + ": " + err.Error(),
		}, "", "  ")
		_, _ = os.Stdout.Write(append(payload, '\n'))
		os.Exit(1)
	}
	log.Fatalf("%s: %v", action, err)
}
