package internal

import (
	"flag"
	"fmt"
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

	cfg.Output.cleanCDNFile()
	cfg.Output.cleanEmbedFile()

	fmt.Print("webdep: cleaned\n")
	return nil
}
