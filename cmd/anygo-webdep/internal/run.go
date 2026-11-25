package internal

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xenc"
	"go.yaml.in/yaml/v4"
)

func init() {
	xcfg.WithDecoder(".yml", xenc.UnmarshalFunc(yaml.Unmarshal))
	xcfg.WithDecoder(".yaml", xenc.UnmarshalFunc(yaml.Unmarshal))
	xcfg.WithDecoder(".toml", xenc.UnmarshalFunc(toml.Unmarshal))
}

func Run(args []string) {
	if len(args) < 2 {
		usage()
		os.Exit(2)
	}

	switch args[1] {
	case "build":
		if err := runBuild(args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "webdep:", err)
			os.Exit(1)
		}
	case "clean":
		if err := runClean(args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "webdep:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`anygo-webdep - web asset build tool

Commands:
  anygo-webdep build [-c webdep.yaml] [-mode cdn|embed]
  anygo-webdep clean [-c webdep.yaml]

https://github.com/xanygo/anygo
`)
}
