//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-02

package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xanygo/anygo/xarchive"
	"github.com/xanygo/anygo/xcodec"
	"github.com/xanygo/anygo/xstr"
)

var outfile = flag.String("o", "out.ez", "output file name")
var token = flag.String("token", "anygo-3000", "token for encryption")

var exeName = "[" + os.Args[0] + "] "

func main() {
	flag.Parse()
	if *outfile == "" {
		log.Fatal("-o flag is required")
	}
	content := createZip()
	rd, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	assert(err, "zip.NewReader")
	names := xarchive.ZipFileNames(rd, 0)

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
}

func getMsgPrefix() string {
	wd, _ := os.Getwd()
	wd = xstr.CutLastNAfter(wd, string(filepath.Separator), 3)
	txt := exeName + fmt.Sprintf("%30s", wd) + " : "
	return txt
}

func createZip() []byte {
	bf := &bytes.Buffer{}
	w := zip.NewWriter(bf)
	for _, fp := range flag.Args() {
		info, err := os.Stat(fp)
		assert(err, "open "+fp)
		if info.IsDir() {
			err = filepath.WalkDir(fp, func(name string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				return oneFile(w, name)
			})
			// err = w.AddFS(os.DirFS(fp))
			assert(err, "AddFS "+fp)
		} else {
			tw, err := w.Create(fp)
			assert(err, "Create "+fp)
			content, err := os.ReadFile(fp)
			assert(err, "ReadFile "+fp)
			_, err = tw.Write(content)
			assert(err, "Write "+fp)
		}
	}
	err := w.Close()
	assert(err, "w.Close")
	return bf.Bytes()
}

func oneFile(w *zip.Writer, name string) error {
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
	fw, err := w.CreateHeader(h)
	if err != nil {
		return err
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(fw, f)
	return err
}

func assert(err error, msg string) {
	if err != nil {
		log.Fatal(msg, err)
	}
}
