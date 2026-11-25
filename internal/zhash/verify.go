package zhash

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// VerifyFile 校验本地文件
func VerifyFile(path string, newHash func() hash.Hash, expected []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := newHash()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	if !bytes.Equal(h.Sum(nil), expected) {
		return errors.New("hash mismatch")
	}

	return nil
}

// ParseHash 解析算法和预期 Hash
func ParseHash(value string, hashFactory func(string) func() hash.Hash) (func() hash.Hash, []byte, error) {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		parts = strings.SplitN(value, "-", 2)
	}
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid hash format: %q", value)
	}
	fnName := strings.ToLower(parts[0])
	newHash := hashFactory(fnName)
	if newHash == nil {
		return nil, nil, fmt.Errorf("unsupported hash algorithm: %s", parts[0])
	}

	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid hash value: %w", err)
	}

	expectSize := newHash().Size()

	if len(expected) != expectSize {
		return nil, nil, fmt.Errorf("invalid %s hash length, expect %d, got %d", parts[0], len(expected), expectSize)
	}

	return newHash, expected, nil
}
