//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-21

package xsmtp

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/xanygo/anygo/ds/xslice"
)

const Protocol = "SMTP"

type Mail struct {
	Subject     string // 邮件标题，必填
	ContentType string // ContentType 正文的类型，可选，默认为 text/html;charset=utf-8
	Content     string // 邮件正文，必填

	To  []string // 接收地址，To、CC、BCC 三者至少有一个有效地址，可以是：name@example.com 或者 别名|name@example.com
	CC  []string // 抄送地址
	BCC []string // 暗抄地址

	// From 发送着，可选，当为空时会使用配置中的 Username 字段（登录 smtp 服务器的邮件地址）
	From string

	// MsgID 邮件ID,可选，可用于邮件去重
	// 如 <20251021134757.674B133CB0@example.com>
	MsgID string

	// Date 邮件的创建时间，可选
	Date time.Time

	// Attachment 提供下载的附件，可选
	Attachment []*Attachment

	// Inline 内联的图片、视频等资源，可选
	Inline []*InlineResource
}

func (req *Mail) check() error {
	if req.Subject == "" {
		return errors.New("empty Subject")
	}
	if len(req.To) == 0 && len(req.BCC) == 0 && len(req.CC) == 0 {
		return errors.New("empty To, BCC, or CC")
	}
	if err := checkAddress(req.To...); err != nil {
		return err
	}
	if err := checkAddress(req.BCC...); err != nil {
		return err
	}
	if err := checkAddress(req.CC...); err != nil {
		return err
	}
	if req.Content == "" {
		return errors.New("empty Content")
	}
	for idx, at := range req.Attachment {
		if err := at.check(); err != nil {
			return fmt.Errorf("attachment[%d]:%w", idx, err)
		}
	}
	for idx, at := range req.Inline {
		if err := at.check(); err != nil {
			return fmt.Errorf("inlineResource[%d]:%w", idx, err)
		}
	}
	return nil
}

func (req *Mail) AddAttachFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("file %q is a directory", path)
	}
	a := &Attachment{
		Name: filepath.Base(path),
		Open: func() (io.Reader, error) {
			return os.Open(path)
		},
	}
	req.Attachment = append(req.Attachment, a)
	return nil
}

func (req *Mail) AddInlineFile(path string, cid string) error {
	if cid == "" {
		return errors.New("empty cid")
	}
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("file %q is a directory", path)
	}
	_, found := xslice.FilterOne(req.Inline, func(index int, item *InlineResource) bool {
		return item.CID == cid
	})
	if found {
		return fmt.Errorf("duplicate inline file %q with same CID %q", path, cid)
	}
	a := &InlineResource{
		Name: filepath.Base(path),
		Open: func() (io.Reader, error) {
			return os.Open(path)
		},
		CID: cid,
	}
	req.Inline = append(req.Inline, a)
	return nil
}

type Attachment struct {
	Name        string                    // 文件名，必填，不得包含目录结构
	ContentType string                    // 类型，可选
	Open        func() (io.Reader, error) // 文件内容的 Reader，必填，若返回的时 io.ReadCloser，读取后会自动 close
}

func (a *Attachment) getContentType() string {
	if a.ContentType == "" {
		ext := filepath.Ext(a.Name)
		if ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				return ct
			}
		}
		return "application/octet-stream"
	}
	return a.ContentType
}

var pureFilenameRegex = regexp.MustCompile(`^[^/\\]+$`)

func (a *Attachment) check() error {
	if a.Name == "" {
		return errors.New("empty Name")
	}
	if !pureFilenameRegex.MatchString(a.Name) {
		return fmt.Errorf("invalid Name: %q", a.Name)
	}
	if a.Open == nil {
		return errors.New("empty Open")
	}
	return nil
}

type InlineResource struct {
	Name        string                    // 文件名，必填，不得包含目录结构
	ContentType string                    // 类型，可选
	Open        func() (io.Reader, error) // 文件内容的 Reader，必填，若返回的时 io.ReadCloser，读取后会自动 close
	CID         string                    // 必填，正文里引用的资源 ID，最好是使用字母，数字的组合
}

func (a *InlineResource) getContentType() string {
	if a.ContentType == "" {
		ext := filepath.Ext(a.Name)
		if ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				return ct
			}
		}
		return "application/octet-stream"
	}
	return a.ContentType
}

func (a *InlineResource) check() error {
	if a.Name == "" {
		return errors.New("empty Name")
	}
	if !pureFilenameRegex.MatchString(a.Name) {
		return fmt.Errorf("invalid Name: %q", a.Name)
	}
	if a.Open == nil {
		return errors.New("empty Open")
	}
	if a.CID == "" {
		return errors.New("empty CID")
	}
	return nil
}
