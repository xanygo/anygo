// Package main 查找 go generate 命令，然后并行执行
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/xanygo/anygo/cli/xcolor"
)

var concurrency = flag.Int("j", 8, "maximum number of concurrent generate tasks")
var tags = flag.String("tags", "", "build tags")
var quiet = flag.Bool("q", false, "quiet, redirect child process stdout and stderr to null")

var limiter chan struct{}

var errs []error
var mux sync.Mutex

var wg sync.WaitGroup

func main() {
	flag.Parse()
	limiter = make(chan struct{}, *concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := flag.Arg(0)
	if dir == "" {
		dir = "."
	}

	err := fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		name := d.Name()
		if d.IsDir() {
			if len(name) > 1 && name[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(name, ".go") || strings.HasPrefix(name, "_") {
			return nil
		}
		tryRunFile(ctx, filepath.Join(dir, path))
		return nil
	})
	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}
	if len(errs) > 0 {
		log.Fatal(errors.Join(errs...))
	}
}

func tryRunFile(ctx context.Context, fp string) {
	src, err := os.ReadFile(fp)
	if err != nil {
		log.Fatalf("generate: %s", err)
	}
	if !bytes.Contains(src, []byte("//go:generate")) {
		return
	}

	// 包含有 go:build 的还是执行 go generate file.go
	if bytes.Contains(src, []byte("//go:build ")) || bytes.Contains(src, []byte("//go:build\t")) {
		if isFileGenerate(src) {
			runGoGenerate(ctx, fp)
		}
	} else {
		tryRunDirect(ctx, fp, src)
	}
}

func runGoGenerate(ctx context.Context, fp string) {
	args := []string{"generate"}
	if *tags != "" {
		args = append(args, "-tags", *tags)
	}
	args = append(args, fp)
	cmd := exec.CommandContext(ctx, "go", args...)
	execCmd(fp, cmd)
}

func execCmd(fp string, cmd *exec.Cmd) {
	log.Println(xcolor.RedString(cmd.Dir), xcolor.GreenString(cmd.String()))
	if *quiet {
		cmd.Stderr = io.Discard
		cmd.Stdout = io.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	wg.Go(func() {
		limiter <- struct{}{}
		defer func() {
			<-limiter
		}()
		err1 := cmd.Run()
		if err1 != nil {
			mux.Lock()
			errs = append(errs, fmt.Errorf("%s: %w", fp, err1))
			mux.Unlock()
		}
	})
}

// 直接运行解析出来的命令
func tryRunDirect(ctx context.Context, fp string, src []byte) {
	cmds := generateCommands(src)
	if len(cmds) == 0 {
		return
	}
	dir := filepath.Dir(fp)
	for _, args := range cmds {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = dir
		execCmd(fp, cmd)
	}
}

func isFileGenerate(src []byte) bool {
	lines := bytes.Split(src, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("//go:build ignore")) || bytes.HasPrefix(line, []byte("//go:build\tignore")) {
			return false
		}
		if isGoGenerate(line) {
			return true
		}
	}
	return false
}

func generateCommands(src []byte) [][]string {
	var result [][]string
	lines := bytes.Split(src, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("//go:build ignore")) || bytes.HasPrefix(line, []byte("//go:build\tignore")) {
			return nil
		}
		if isGoGenerate(line) {
			ws, err := parserCommandLine(string(line))
			if err != nil {
				log.Fatal(err)
			}
			if len(ws) > 0 {
				result = append(result, ws)
			}
		}
	}
	return result
}

var kw1 = []byte("//go:generate ")
var kw2 = []byte("//go:generate\t")

func isGoGenerate(buf []byte) bool {
	return bytes.HasPrefix(buf, kw1) || bytes.HasPrefix(buf, kw2)
}

func parserCommandLine(line string) ([]string, error) {
	raw := line
	var words []string
	line = line[len(kw1):]
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
Words:
	for {
		line = strings.TrimLeft(line, " \t")
		if len(line) == 0 {
			break
		}
		if line[0] == '"' {
			for i := 1; i < len(line); i++ {
				c := line[i] // Only looking for ASCII so this is OK.
				switch c {
				case '\\':
					if i+1 == len(line) {
						return nil, fmt.Errorf("bad backslash: %q", raw)
					}
					i++ // Absorb next byte (If it's a multibyte we'll get an error in Unquote).
				case '"':
					word, err := strconv.Unquote(line[0 : i+1])
					if err != nil {
						return nil, fmt.Errorf("bad quoted string: %q", raw)
					}
					words = append(words, word)
					line = line[i+1:]
					// Check the next character is space or end of line.
					if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
						return nil, fmt.Errorf("expect space after quoted argument: %q", raw)
					}
					continue Words
				}
			}
			return nil, fmt.Errorf("mismatched quoted string: %q", raw)
		}
		i := strings.IndexAny(line, " \t")
		if i < 0 {
			i = len(line)
		}
		words = append(words, line[0:i])
		line = line[i:]
	}
	return words, nil
}
