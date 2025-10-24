//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-24

package xsmtp

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/smtp"
	"time"

	"github.com/xanygo/anygo/ds/xmetric"
	"github.com/xanygo/anygo/ds/xoption"
	"github.com/xanygo/anygo/xnet"
	"github.com/xanygo/anygo/xnet/xrpc"
)

var _ xrpc.Request = (*request)(nil)

type request struct {
	mails iter.Seq[*Mail]
}

func (r request) String() string {
	return "smtp request"
}

func (r request) Protocol() string {
	return Protocol
}

func (r request) APIName() string {
	return "Send"
}

func (r request) WriteTo(ctx context.Context, w *xnet.ConnNode, opt xoption.Reader) error {
	cr, ok := w.Handshake.(*handshakeReply)
	if !ok {
		return fmt.Errorf("invalid handshake type: %T, not smtp client", w.Handshake)
	}
	ctx, span := xmetric.Start(ctx, Protocol)
	var cnt int
	defer func() {
		cr.client.Quit()
		span.SetAttributes(xmetric.AnyAttr("sent", cnt))
		span.End()
	}()

	for m := range r.mails {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}
		if err := sendOne(w, opt, cr.client, cr.username, m); err != nil {
			return err
		}
		cnt++
	}
	return nil
}

// sendOne 发送邮件，这里的 client 是以及登录完成的
func sendOne(conn *xnet.ConnNode, opt xoption.Reader, client *smtp.Client, username string, m *Mail) error {
	if err := m.check(); err != nil {
		return err
	}

	totalTimeout := xoption.WriteReadTimeout(opt)
	if err := conn.SetDeadline(time.Now().Add(totalTimeout)); err != nil {
		return err
	}
	defer conn.SetDeadline(time.Time{})

	from := m.From
	if from == "" {
		from = username
	}
	if from == "" {
		return errors.New("empty From")
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := setRcpt(client, m.To...); err != nil {
		return err
	}
	if err := setRcpt(client, m.CC...); err != nil {
		return err
	}
	if err := setRcpt(client, m.BCC...); err != nil {
		return err
	}
	wc, err := client.Data()
	if err != nil {
		return err
	}

	w1 := &writer{
		w: wc,
	}
	w1.to("To", m.To...)
	w1.to("Cc", m.CC...)

	w1.subject(m.Subject)
	if m.MsgID != "" {
		w1.simpleHeader("Message-Id", m.MsgID)
	}
	if m.Date.IsZero() {
		w1.simpleHeader("Date", time.Now().Format(time.RFC1123Z))
	} else {
		w1.simpleHeader("Date", m.Date.Format(time.RFC1123Z))
	}
	w1.simpleHeader("MIME-Version", "1.0")

	if len(m.Attachment) > 0 || len(m.Inline) > 0 {
		w1.multipart(m.ContentType, m.Content, m.Attachment, m.Inline)
	} else {
		w1.content(m.ContentType, m.Content)
	}

	err = w1.err
	if err == nil {
		err = wc.Close()
	}
	return err
}
