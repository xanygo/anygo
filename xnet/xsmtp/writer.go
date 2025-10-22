//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-21

package xsmtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/xanygo/anygo/ds/xstr"
	"github.com/xanygo/anygo/xio"
)

type writer struct {
	w   io.Writer
	n   int64
	err error
}

func (w *writer) saveCntError(n int64, err error) {
	w.n += n
	if w.err == nil && err != nil {
		w.err = err
	}
}

func (w *writer) writeString(s string) {
	n, err := io.WriteString(w.w, s)
	w.saveCntError(int64(n), err)
}

func (w *writer) simpleHeader(key string, value string) {
	w.writeString(key)
	w.writeString(": ")
	w.writeString(value)
	w.writeString(crlf)
}

func (w *writer) subject(subject string) {
	w.simpleHeader("Subject", encodeRFC2047(subject))
}

func (w *writer) content(ct string, content string) {
	if ct != "" {
		w.simpleHeader("Content-Type", ct)
	} else {
		w.simpleHeader("Content-Type", "text/html; charset=utf-8")
	}
	w.simpleHeader("Content-Transfer-Encoding", "quoted-printable")
	w.writeString(crlf)
	qw := quotedprintable.NewWriter(w.w)
	n, err := qw.Write([]byte(content))
	w.saveCntError(int64(n), err)
}

func (w *writer) multipart(contentType, content string, ats []*Attachment, inline []*InlineResource) {
	boundary := "mix_" + xstr.RandNChar(16)
	w.simpleHeader("Content-Type", `multipart/mixed; boundary="`+boundary+`"`)
	w.writeString(crlf)

	mw := multipart.NewWriter(w.w)
	mw.SetBoundary(boundary)

	if len(inline) > 0 {
		relatedBoundary := "rel_" + xstr.RandNChar(16)
		h1 := textproto.MIMEHeader{}
		h1.Set("Content-Type", fmt.Sprintf("multipart/related; boundary=%q", relatedBoundary))
		pw1, err := mw.CreatePart(h1)
		w.saveCntError(0, err)
		if w.err != nil {
			return
		}

		hw := multipart.NewWriter(pw1)
		hw.SetBoundary(relatedBoundary)

		w.multipartContent(hw, contentType, content)

		for _, res := range inline {
			w.inlineResource(hw, res)
		}
		w.saveCntError(0, hw.Close())
	} else {
		w.multipartContent(mw, contentType, content)
	}
	for _, at := range ats {
		w.attachment(mw, at)
	}
	w.saveCntError(0, mw.Close())
}

func (w *writer) multipartContent(mw *multipart.Writer, contentType, content string) {
	altBoundary := "alt_" + xstr.RandNChar(16)
	h0 := textproto.MIMEHeader{}
	h0.Set("Content-Type", "multipart/alternative; boundary="+fmt.Sprintf("%q", altBoundary))
	pw0, err := mw.CreatePart(h0)
	w.saveCntError(0, err)
	if w.err != nil {
		return
	}

	zw := multipart.NewWriter(pw0)
	zw.SetBoundary(altBoundary)

	h := textproto.MIMEHeader{}
	if contentType == "" {
		h.Set("Content-Type", "text/html; charset=utf-8")
	} else {
		h.Set("Content-Type", contentType)
	}
	h.Set("Content-Transfer-Encoding", "base64")
	pw, err := zw.CreatePart(h)
	w.saveCntError(0, err)
	if w.err != nil {
		return
	}
	w.writeBase64(pw, bytes.NewBufferString(content))
	w.saveCntError(0, zw.Close())
}

func (w *writer) attachment(mw *multipart.Writer, a *Attachment) {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", fmt.Sprintf(a.getContentType()+"; name=%q", a.Name))
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", a.Name))
	h.Set("Content-Transfer-Encoding", "base64")
	pw, err := mw.CreatePart(h)
	w.saveCntError(0, err)
	if w.err != nil {
		return
	}
	rd, err := a.Open()
	if err != nil {
		w.saveCntError(0, err)
		return
	}
	if rc, ok := rd.(io.Closer); ok {
		defer rc.Close()
	}
	w.writeBase64(pw, rd)
}

func (w *writer) inlineResource(mw *multipart.Writer, a *InlineResource) {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", a.getContentType()+fmt.Sprintf("; name=%q", a.Name))
	h.Set("Content-ID", "<"+a.CID+">")
	h.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", a.Name))
	h.Set("Content-Transfer-Encoding", "base64")
	pw, err := mw.CreatePart(h)
	w.saveCntError(0, err)
	if w.err != nil {
		return
	}
	rd, err := a.Open()
	if err != nil {
		w.saveCntError(0, err)
		return
	}
	if rc, ok := rd.(io.Closer); ok {
		defer rc.Close()
	}
	w.writeBase64(pw, rd)
}

func (w *writer) writeBase64(to io.Writer, rd io.Reader) {
	lw := &xio.LineLengthWriter{
		W:      to,
		MaxLen: 76,
		Sep:    []byte(crlf),
	}
	bw := base64.NewEncoder(base64.StdEncoding, lw)
	n, err := io.Copy(bw, rd)
	w.saveCntError(n, err)
}

func (w *writer) to(key string, addressList ...string) {
	if len(addressList) == 0 {
		return
	}
	w.writeString(key)
	w.writeString(": ")
	for idx, to := range addressList {
		to = strings.TrimSpace(to)
		name, addr, found := strings.Cut(to, "|")
		if found {
			w.writeString(encodeRFC2047(name))
			w.writeString("<")
			w.writeString(addr)
			w.writeString(">")
		} else {
			w.writeString("<" + to + ">")
		}
		if idx < len(addressList)-1 {
			w.writeString(",")
		}
		w.writeString(crlf)
	}
}
