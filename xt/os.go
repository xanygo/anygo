// Copyright(C) 2024 github.com/fsgo  All Rights Reserved.
// Author: hidu <duv123@gmail.com>
// Date: 2024/7/7

package xt

import (
	"errors"
	"io/fs"
	"os"
)

func FileExists(t Testing, path string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("unable to find file %q", path)
		} else {
			t.Fatalf("error when running os.Lstat(%q): %s", path, err)
		}
		return
	}
	if info.IsDir() {
		t.Fatalf("%q is a directory", path)
	}
}

func FileNotExists(t Testing, path string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	info, err := os.Lstat(path)
	if err != nil || info.IsDir() {
		return
	}
	t.Fatalf("file %q exists", path)
}

func DirExists(t Testing, path string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("unable to find file %q", path)
		} else {
			t.Fatalf("error when running os.Lstat(%q): %s", path, err)
		}
		return
	}
	if !info.IsDir() {
		t.Fatalf("%q is a file", path)
	}
}

func DirNotExists(t Testing, path string) {
	if h, ok := t.(Helper); ok {
		h.Helper()
	}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() {
		return
	}
	t.Fatalf("directory %q exists", path)
}
