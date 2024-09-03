//  Copyright(C) 2024 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2024-09-03

package hook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/xanygo/anygo/xattr"
	"github.com/xanygo/anygo/xcfg/internal/parser"
	"github.com/xanygo/anygo/xcfg/internal/zcache"
)

const hookTplPrefix = "hook.template "

type Template struct {
}

func (t *Template) Hook(ctx context.Context, cfPath string, content []byte) ([]byte, error) {
	cmts := parser.HeadComments(content)
	if len(cmts) == 0 {
		return content, nil
	}
	params := make(map[string]string, 3)
	for _, cmt := range cmts {
		if strings.HasPrefix(cmt, hookTplPrefix) {
			arr := strings.Fields(cmt[len(hookTplPrefix):])
			for i := 0; i < len(arr); i++ {
				tmp := strings.Split(arr[i], "=")
				if len(tmp) == 2 && len(tmp[0]) > 0 && len(tmp[1]) > 0 {
					params[tmp[0]] = tmp[1]
				}
			}
		}
	}
	if params["Enable"] != "true" {
		return content, nil
	}
	return t.exec(ctx, cfPath, content, params)
}

func (t *Template) exec(ctx context.Context, cfPath string, content []byte, tp map[string]string) (output []byte, err error) {
	tmpl := template.New("config")
	left := "{{"
	right := "}}"
	if v := tp["Left"]; len(v) > 0 {
		left = v
	}
	if v := tp["Right"]; len(v) > 0 {
		right = v
	}
	tmpl.Delims(left, right)
	tmpl.Funcs(map[string]any{
		"include": func(name string) (string, error) {
			return t.fnInclude(ctx, name, cfPath, tp)
		},
		"fetch": func(name string, args ...string) (string, error) {
			return t.fnFetch(ctx, tp, name, args)
		},
		"osenv": func(name string) string {
			return os.Getenv(name)
		},
		"contains": func(s string, sub string) bool {
			return strings.Contains(s, sub)
		},
		"prefix": func(s string, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"suffix": func(s string, suffix string) bool {
			return strings.HasSuffix(s, suffix)
		},
	})
	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}

	data := map[string]string{
		"IDC":         xattr.IDC(),
		"RootDir":     xattr.RootDir(),
		"ConfRootDir": xattr.ConfDir(),
		"LogRootDir":  xattr.LogDir(),
		"DataRootDir": xattr.DataDir(),
		"RunMode":     xattr.RunMode().String(),
	}

	if err = tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *Template) pathHasMeta(path string) bool {
	magicChars := `*?[`
	if runtime.GOOS != "windows" {
		magicChars = `*?[\`
	}
	return strings.ContainsAny(path, magicChars)
}

func (t *Template) fnInclude(ctx context.Context, name string, cfPath string, tp map[string]string) (string, error) {
	if cfPath == "" {
		return "", errors.New("config's FilePath is empty cannot use include")
	}
	var fp string
	if filepath.IsAbs(name) {
		fp = name
	} else {
		fp = filepath.Join(filepath.Dir(cfPath), name)
	}

	files, err := filepath.Glob(fp)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		if !t.pathHasMeta(name) {
			return "", fmt.Errorf("include %q not found", name)
		}
		return "", nil
	}
	var buf bytes.Buffer
	for _, f := range files {
		body, err1 := os.ReadFile(f)
		if err1 != nil {
			return "", err1
		}

		o1, err2 := t.exec(ctx, f, body, tp)
		if err2 != nil {
			return "", err2
		}
		buf.Write(o1)
	}
	return buf.String(), nil
}

func (t *Template) getXCache() *zcache.FileCache {
	dir := filepath.Join(xattr.TempDir(), "xcfg_cache")
	return &zcache.FileCache{
		Dir: dir,
	}
}

func (t *Template) fnFetch(ctx context.Context, tp map[string]string, api string, ps []string) (string, error) {
	if len(api) == 0 {
		return "", errors.New("url is required")
	}
	if len(ps) > 1 {
		return "", errors.New("only support 0 or 1 param")
	}

	timeout := 3 * time.Second
	var cacheTTL time.Duration
	if len(ps) == 1 {
		param, err := zcache.ParserParam(ps[0])
		if err != nil {
			return "", err
		}
		if param.Timeout > 0 {
			timeout = param.Timeout
		}
		cacheTTL = param.TTL
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	bf, err := httpFetch(ctx, api)

	if cacheTTL > 0 {
		fc := t.getXCache()
		if err == nil {
			fc.Set(api, bf)
		} else {
			if cv, ok := fc.Get(api, cacheTTL); ok {
				return string(cv), nil
			}
		}
	}

	return string(bf), err
}

func httpFetch(ctx context.Context, api string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
