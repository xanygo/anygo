//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-02

package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/xanygo/anygo/ds/xhash"
	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/ds/xzip"
	"github.com/xanygo/anygo/xcodec"
)

var outfile = flag.String("o", "asset.ez", "output file name")
var token = flag.String("token", "anygo-3000", "token for encryption")

var minify = flag.String("m", "", `minify files with the specified file extension.
https://github.com/tdewolff/minify/tree/master/cmd/minify

e.g.:  js,css
`)

var goFile = flag.String("go", "", "generated go file name, e.g. asset_ez.go")
var goVar = flag.String("var", "asset", "go variable name")
var goPkg = flag.String("pkg", "", "go package name")
var tags = flag.String("tags", "release", "for generated go file //go:build {tags}")

var exeName = "[" + os.Args[0] + "] "

func main() {
	flag.Parse()
	if *outfile == "" {
		log.Fatal("-o flag is required")
	}
	content := createZip()
	rd, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	assert(err, "zip.NewReader")
	names := xzip.FileNames(rd, 0)

	ez := &xcodec.AesOFB{
		Key: *token,
	}
	ct, err := ez.Encrypt(content)
	assert(err, "encrypt content")

	xm := md5.New()
	xm.Write(ct)
	xm.Write(ez.ID())
	sign := xm.Sum(nil)

	prefix := getMsgPrefix()

	old, _ := os.ReadFile(*outfile)
	if len(old) > 32 && bytes.Equal(old[len(old)-16:], sign) {
		fmt.Fprintf(os.Stderr, "%s %-15s %s\n", prefix, *outfile, "not changed")
		return
	}

	file, err := os.Create(*outfile)
	assert(err, "create output file: "+*outfile)
	defer file.Close()
	_, err = file.Write(ct)
	assert(err, "write to output file")

	_, err = file.Write(sign)
	assert(err, "write sign output file")

	kb := fmt.Sprintf("%.2f", float64(len(ct)+len(sign))/1024.0)
	fmt.Fprintln(os.Stderr, prefix, flag.Args(), "->", *outfile, kb, "kb", len(names), "files")
	space := strings.Repeat(" ", len(exeName))
	for idx, name := range names {
		fmt.Fprintf(os.Stderr, "%s %03d    %s \n", space, idx, name)
	}

	generateGoFile()
}

func getMsgPrefix() string {
	wd, _ := os.Getwd()
	wd = xstr.CutLastNAfter(wd, string(filepath.Separator), 3)
	txt := exeName + fmt.Sprintf("%30s", wd) + " : "
	return txt
}

func createZip() []byte {
	bf := &bytes.Buffer{}
	zw := zip.NewWriter(bf)
	zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	for _, fp := range flag.Args() {
		info, err := os.Stat(fp)
		assert(err, "open "+fp)
		if info.IsDir() {
			err = filepath.WalkDir(fp, func(name string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				return oneFile(zw, name)
			})
			// err = w.AddFS(os.DirFS(fp))
			assert(err, "AddFS "+fp)
		} else {
			tw, err := zw.Create(fp)
			assert(err, "Create "+fp)
			content, err := os.ReadFile(fp)
			assert(err, "ReadFile "+fp)
			_, err = tw.Write(content)
			assert(err, "Write "+fp)
		}
	}
	err := zw.Close()
	assert(err, "w.Close")
	return bf.Bytes()
}

func oneFile(zw *zip.Writer, name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("zip: cannot add non-regular file")
	}
	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	h.Name = name
	h.Method = zip.Deflate
	fw, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	content, err := readAndMinify(name)
	if err != nil {
		return err
	}
	_, err = fw.Write(content)
	return err
}

func assert(err error, msg string) {
	if err != nil {
		log.Fatal(msg, err)
	}
}

// https://github.com/tdewolff/minify/tree/master/cmd/minify
// Ubuntu: sudo apt install minify
func readAndMinify(name string) ([]byte, error) {
	baseName := filepath.Base(name)
	if strings.Contains(baseName, ".min.") {
		return os.ReadFile(name)
	}
	ext := strings.TrimPrefix(filepath.Ext(baseName), ".")
	ms := strings.Split(*minify, ",")
	if !slices.Contains(ms, ext) {
		return os.ReadFile(name)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "minify", name)
	return cmd.Output()
}

func generateGoFile() {
	fileName := *goFile
	if !strings.HasSuffix(fileName, ".go") {
		log.Println("skipped generating go file")
		return
	}
	pkg := *goPkg
	if pkg == "" {
		fp, err := filepath.Abs(*goFile)
		assert(err, "filepath.Abs")
		pkg = filepath.Base(filepath.Dir(fp))
	}

	vars := map[string]any{
		"PKG":    pkg,
		"Var":    *goVar,
		"VarTmp": "_" + xhash.Md5(*outfile+"|"+*goVar+"|"+pkg),
		"Token":  tokenStr(*token),
		"EZFile": *outfile,
		"Tags":   *tags,
	}
	str, err := xstr.RenderTemplate(goFileTpl, vars)
	assert(err, "Render Go file: "+fileName)

	str = strings.TrimSpace(str)
	bf, err := format.Source([]byte(str))
	assert(err, "Format Go file: "+fileName)

	err = os.WriteFile(fileName, bf, 0644)
	assert(err, "Write Go file")

	space := strings.Repeat(" ", len(exeName))
	fmt.Fprintln(os.Stderr, space+" generated "+fileName)
}

func tokenStr(token string) string {
	// string([]byte{'a'})
	var b strings.Builder
	b.WriteString("string([]byte{\n")
	for i, c := range token {
		b.WriteString(fmt.Sprintf("'%c'", c))
		b.WriteString(",")
		if (i+1)%8 == 0 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n})")
	return b.String()
}

const goFileTpl = `
// Code generated by anygo-encrypt-zip DO NOT EDIT.

{{ if .Tags }}
//go:build {{ .Tags }}
{{ end }}

package {{.PKG}}

import (
	_ "embed"
	"io/fs"

	"github.com/xanygo/anygo/ds/xzip"
	"github.com/xanygo/anygo/xcodec"
)

var {{.Var}} fs.FS

//go:embed {{.EZFile}}
var {{.VarTmp}} []byte

func init() {
	dz := &xcodec.AesOFB{
		Key:{{.Token}},
	}
	{{.Var}} = xzip.MustDecrypt({{.VarTmp}}, dz)
	clear({{.VarTmp}})
	{{.VarTmp}} = nil
}
`
