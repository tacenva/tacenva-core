//go:build windows

package config

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	moveFileEx = kernel32.NewProc("MoveFileExW")
)

const (
	moveFileReplaceExisting = 0x1
	moveFileWriteThrough    = 0x8
)

func atomicWrite(
	path string,
	data []byte,
	perm os.FileMode,
) error {
	tempFile, err := os.CreateTemp(
		filepath.Dir(path),
		".config-*.tmp",
	)
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if err := tempFile.Chmod(perm); err != nil {
		tempFile.Close()
		return err
	}

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	from, err := syscall.UTF16PtrFromString(tempPath)
	if err != nil {
		return err
	}

	to, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	result, _, callErr := moveFileEx.Call(
		uintptr(unsafe.Pointer(from)),
		uintptr(unsafe.Pointer(to)),
		uintptr(
			moveFileReplaceExisting|
				moveFileWriteThrough,
		),
	)

	if result == 0 {
		return callErr
	}

	return nil
}
