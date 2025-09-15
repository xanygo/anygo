//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-09-15

package xoption

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	KeyTLSConfig = NewKey("tls.Config")
)

func SetTLSConfig(opt Writer, c *tls.Config) {
	opt.Set(KeyTLSConfig, c)
}

func GetTLSConfig(opt Reader) *tls.Config {
	return GetAsDefault[*tls.Config](opt, KeyTLSConfig, nil)
}

type TLSConfig struct {
	SkipVerify bool   `json:"SkipVerify" yaml:"SkipVerify"`
	ServerName string `json:"ServerName" yaml:"ServerName"`

	// CAFile 根证书（CA），用于信任自签名证书,如   ca.crt
	CAFile string `json:"CAFile" yaml:"CAFile"`

	// CertFile 客户端证书,如"client.crt"
	CertFile string `json:"CertFile" yaml:"CertFile"`

	// KeyFile 客户端证私钥，如  client.key
	KeyFile string `json:"KeyFile" yaml:"KeyFile"`
}

func (c *TLSConfig) readPEMorFile(data string) ([]byte, error) {
	if strings.HasPrefix(data, "-----BEGIN") {
		return []byte(data), nil // 直接是 PEM 内容
	}
	return os.ReadFile(data) // 当文件路径
}

func (c *TLSConfig) Parser() (*tls.Config, error) {
	if c == nil {
		return nil, nil
	}
	tc := &tls.Config{
		InsecureSkipVerify: c.SkipVerify,
		ServerName:         c.ServerName,
	}
	c.CAFile = strings.TrimSpace(c.CAFile)
	if c.CAFile != "" {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system CA pool: %w", err)
		}
		caCert, err := c.readPEMorFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("append CA cert failed")
		}
		tc.RootCAs = caCertPool
	}

	c.CertFile = strings.TrimSpace(c.CertFile)
	c.KeyFile = strings.TrimSpace(c.KeyFile)

	// 如果指定了客户端证书
	if c.CertFile != "" && c.KeyFile != "" {
		certPEM, err := c.readPEMorFile(c.CertFile)
		if err != nil {
			return nil, fmt.Errorf("load cert: %w", err)
		}
		keyPEM, err := c.readPEMorFile(c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load key: %w", err)
		}
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("parser cert/key: %w", err)
		}
		tc.Certificates = []tls.Certificate{cert}
	}

	return tc, nil
}
