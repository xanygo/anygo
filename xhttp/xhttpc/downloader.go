package xhttpc

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xanygo/anygo/internal/zhash"
)

type Downloader struct {
	URL         string // 必填，url 地址
	HashValue   string // 可选，校验的hash 地址，如 sha256:xxx
	HashFactory func(string) func() hash.Hash
	OutputPath  string   // 必填，本地地址，若本地文件存在，并且和 Hash 一致，则不下载
	Client      Executor // 可选，为空时使用默认的 http.Client
	MaxSize     int64    // 可选，最大文件尺寸
}

func (d *Downloader) getClient() Executor {
	if d.Client != nil {
		return d.Client
	}
	return DefaultClient
}

func (d *Downloader) hashFactory(name string) func() hash.Hash {
	if d.HashFactory != nil {
		return d.HashFactory(name)
	}
	switch name {
	case "sha256":
		return sha256.New
	case "sha512":
		return sha512.New
	case "md5":
		return md5.New
	default:
		return nil
	}
}

func (d *Downloader) Run(ctx context.Context) error {
	if d.URL == "" {
		return errors.New("downloader: URL is empty")
	}
	if d.OutputPath == "" {
		return errors.New("downloader: OutputPath is empty")
	}

	var newHash func() hash.Hash
	var expected []byte
	if d.HashValue != "" {
		var err error
		newHash, expected, err = zhash.ParseHash(d.HashValue, d.hashFactory)
		if err != nil {
			return err
		}

		// 检查本地文件是否已经存在且校验通过。
		if err := zhash.VerifyFile(d.OutputPath, newHash, expected); err == nil {
			return nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			// 校验不一致时允许重新下载。
			if _, statErr := os.Stat(d.OutputPath); statErr == nil {
				// 文件存在，但 Hash 不匹配，继续下载。
			} else if !errors.Is(statErr, fs.ErrNotExist) {
				return statErr
			}
		}
	}

	// 创建输出目录。
	dir := filepath.Dir(d.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// 创建同目录临时文件，保证最终可以通过 Rename 替换。
	tmp, err := os.CreateTemp(dir, ".download-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	// 发起 HTTP 请求。
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err != nil {
		tmp.Close()
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := d.getClient().Do(req)
	if err != nil {
		tmp.Close()
		return fmt.Errorf("download %q: %w", d.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		tmp.Close()
		return fmt.Errorf("download %q: unexpected HTTP status %s", d.URL, resp.Status)
	}

	var fw io.Writer = tmp
	var h hash.Hash
	if newHash != nil {
		h = newHash()
		// 下载过程中同时计算 Hash。
		fw = io.MultiWriter(tmp, h)
	}

	var rd io.Reader = resp.Body
	if d.MaxSize > 0 {
		rd = io.LimitReader(resp.Body, d.MaxSize)
	}

	if _, err := io.Copy(fw, rd); err != nil {
		tmp.Close()
		return fmt.Errorf("write downloaded file: %w", err)
	}

	if h != nil && !bytes.Equal(h.Sum(nil), expected) {
		tmp.Close()
		return fmt.Errorf("download %q: hash mismatch", d.URL)
	}

	// 确保数据写入磁盘。
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync downloaded file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close downloaded file: %w", err)
	}

	if err := os.Remove(d.OutputPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove old output: %w", err)
	}

	if err := os.Rename(tmpPath, d.OutputPath); err != nil {
		return fmt.Errorf("rename downloaded file: %w", err)
	}

	return nil
}
