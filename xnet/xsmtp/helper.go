//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-21

package xsmtp

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strings"
)

const crlf = "\r\n"

// encodeRFC2047 将输入字符串编码为符合 RFC 2047 的 UTF-8 Base64 邮件头。
// 自动拆分为多个 encoded-word，每个不超过 75 字符。
func encodeRFC2047(s string) string {
	const (
		charset  = "UTF-8"
		encoding = "B"
		maxLen   = 75 // RFC 2047: encoded-word 最长 75 字符
	)

	prefix := "=?" + charset + "?" + encoding + "?"
	suffix := "?="

	// 计算 Base64 可用长度（去掉前后固定部分长度）
	headerOverhead := len(prefix) + len(suffix)
	maxBase64Len := maxLen - headerOverhead

	// Base64 编码整个字符串（不要先拆分 UTF-8）
	b64 := base64.StdEncoding.EncodeToString([]byte(s))

	// 按长度切割 base64 字符串
	var encodedWords []string
	for len(b64) > 0 {
		n := maxBase64Len
		if len(b64) < n {
			n = len(b64)
		}
		part := b64[:n]
		b64 = b64[n:]
		encodedWords = append(encodedWords, prefix+part+suffix)
	}

	// 用空格分隔 encoded-word（符合 RFC 2047 要求）
	return strings.Join(encodedWords, crlf+" ")
}

func cleanAddress(addr string) string {
	_, to, found := strings.Cut(addr, "|")
	if found {
		return strings.TrimSpace(to)
	}
	return strings.TrimSpace(addr)
}

func setRcpt(client *smtp.Client, list ...string) error {
	for _, to := range list {
		to = cleanAddress(to)
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	return nil
}

var mailReg = regexp.MustCompile(`^\S+@\S+$`)

func checkAddress(list ...string) error {
	for _, to := range list {
		log.Println("checkAddress:", to)
		if strings.Count(to, "@") != 1 {
			return fmt.Errorf("invalid address %q", to)
		}
		to = cleanAddress(to)
		if !mailReg.MatchString(to) {
			return fmt.Errorf("invalid address %q", to)
		}
	}
	return nil
}
