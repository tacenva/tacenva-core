//go:build !windows

package config

import (
	"os"
	"path/filepath"
)

func atomicWrite(
	path string,
	data []byte,
	perm os.FileMode,
) error {
	dir := filepathDir(path)

	tempFile, err := os.CreateTemp(
		dir,
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

	return os.Rename(tempPath, path)
}

func filepathDir(path string) string {
	return filepath.Dir(path)
}
