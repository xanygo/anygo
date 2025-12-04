//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-25

package xfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// NewMover 创建一个配置了默认值的新 Mover 实例。
func NewMover() *Mover {
	return &Mover{
		ChunkSize:     1024 * 1024, // 1MB
		PreservePerms: true,
	}
}

var mover = NewMover()

// Rename 支持跨盘位文件和目录移动操作
func Rename(src, dst string) error {
	return mover.Rename(src, dst)
}

// Mover 支持跨盘位文件和目录移动操作
type Mover struct {
	// ChunkSize 指定复制操作的缓冲区大小。
	ChunkSize int64

	// PreservePerms 标识是否尝试保留文件权限和时间戳。
	PreservePerms bool
}

// Rename 智能地将源路径 (src) 的文件或目录移动到目标路径 (dst)。
//
// 流程：尝试 os.Rename -> 失败且为跨设备错误 -> 执行 copyDir/copyFile -> 删除源
func (m *Mover) Rename(src, dst string) error {
	// 确保源路径存在
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("source path does not exist: %s", src)
	}

	// 1. 尝试原子移动 (Rename)
	// 如果 src 和 dst 在同一文件系统上，os.Rename 将成功。
	err := os.Rename(src, dst)

	if err == nil {
		// 成功原子移动，直接返回
		return nil
	}

	// 2. 判断是否是跨设备错误 (External Device error - EXDEV)
	// 如果不是跨设备错误，但 os.Rename 失败了，直接返回错误
	if !m.isCrossDeviceError(err) {
		return err
	}

	// 3. 跨文件系统移动：复制 -> 删除
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 复制操作
	if srcInfo.IsDir() {
		err = m.copyDir(src, dst)
	} else {
		err = m.copyFile(src, dst)
	}

	if err != nil {
		return fmt.Errorf("failed to copy %s to %s after rename failed: %w", src, dst, err)
	}

	// 复制成功后，删除源文件/目录
	if err = os.RemoveAll(src); err != nil {
		return fmt.Errorf("copy succeeded, but failed to remove source %s: %w", src, err)
	}
	return nil
}

func (m *Mover) isCrossDeviceError(err error) bool {
	if errors.Is(err, syscall.EXDEV) {
		return true
	}
	z := syscall.Errno(17) // for windows
	return errors.Is(err, z)
}

// copyDir 递归地复制目录。
func (m *Mover) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = m.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err = m.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	// 复制目录权限和时间戳
	if m.PreservePerms {
		// 忽略权限和时间戳设置失败的非致命错误
		_ = os.Chmod(dst, srcInfo.Mode())
		_ = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}

	return nil
}

// copyFile 执行文件复制，优先尝试零拷贝，失败则回退。
func (m *Mover) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	totalSize := srcInfo.Size()

	var written int64

	//  标准 I/O 复制 (跨平台回退)
	if m.ChunkSize > 0 {
		buf := make([]byte, m.ChunkSize)
		written, err = io.CopyBuffer(dstFile, srcFile, buf)
	} else {
		written, err = io.Copy(dstFile, srcFile)
	}

	if err != nil {
		return err
	}
	if written != totalSize {
		return fmt.Errorf("copied bytes mismatch, expected %d, got %d", totalSize, written)
	}

	// 设置权限和时间戳
	if m.PreservePerms {
		// 忽略权限和时间戳设置失败的非致命错误
		_ = os.Chmod(dst, srcInfo.Mode())
		_ = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}

	// 确保所有数据写入磁盘
	return dstFile.Sync()
}
