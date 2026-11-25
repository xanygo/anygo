package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xcfg"
	"github.com/xanygo/anygo/xenc"
	"github.com/xanygo/anygo/xhttp/xhttpc"
	"github.com/xanygo/anygo/xlog"
	"github.com/xanygo/anygo/xnet/xrpc"
	"go.yaml.in/yaml/v4"
)

func init() {
	xcfg.WithDecoder(".yml", xenc.UnmarshalFunc(yaml.Unmarshal))
	xcfg.WithDecoder(".yaml", xenc.UnmarshalFunc(yaml.Unmarshal))
	xcfg.WithDecoder(".toml", xenc.UnmarshalFunc(toml.Unmarshal))
}

func setupDebug() {
	xattr.SetRunMode(xattr.ModeDebug)
	lg := xlog.NewSimple(os.Stderr)
	xlog.SetClientLogger(lg)
	xrpc.RegisterIT((&xrpc.Logger{}).Interceptor())
}

func Run(args []string) {
	defer cleanOldTemp()

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

func tmpDir() string {
	return filepath.Join(os.TempDir(), "anygo-webdep")
}

func download(ctx context.Context, url string, hash string) (string, error) {
	sum := md5.Sum([]byte(url))
	m5 := hex.EncodeToString(sum[:])
	tmpPath := filepath.Join(tmpDir(), m5+"_"+path.Base(url))
	log.Println("download temp", filepath.ToSlash(tmpPath))
	dl := &xhttpc.Downloader{
		URL:        url,
		HashValue:  hash,
		OutputPath: tmpPath,
		MaxSize:    maxDownload,
	}
	err := dl.Run(ctx)
	return tmpPath, err
}

func cleanOldTemp() {
	const maxAge = 30 * 24 * time.Hour
	expire := time.Now().Add(-maxAge)
	dir := tmpDir()
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			log.Println("delete ignore, isDir", dir+"/"+path)
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.ModTime().Before(expire) {
			err := os.Remove(path)
			log.Println("delete ", dir+"/"+path, "err=", err)
		}

		return nil
	})
}
