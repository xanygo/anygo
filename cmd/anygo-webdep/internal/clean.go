package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func runClean(args []string) error {
	flags := flag.NewFlagSet("clean", flag.ContinueOnError)
	configPath := flags.String("c", defaultConfig, "configuration file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if cfg.Output.Dir == "" {
		return errors.New("output.dir is required")
	}
	outDir := cfg.Output.Dir
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("remove output directory: %w", err)
	}
	fmt.Printf("webdep: cleaned %s\n", outDir)
	return nil
}
