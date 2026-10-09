// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"os"
)

// FileExists reports whether the named file exists
func FileExists(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}

	return false
}

// DirExists reports whether the dir exists
func DirExists(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		if fi.Mode().IsDir() {
			return true
		}
	}

	return false
}

// EnsureDir ensures that directory exists
func EnsureDir(dirName string, mode int) error {
	err := os.MkdirAll(dirName, os.FileMode(mode))
	if err == nil || os.IsExist(err) {
		return nil
	}

	return err
}

// DeleteDir deletes a dir
func DeleteDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}

	return nil
}
