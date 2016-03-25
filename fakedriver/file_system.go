package fakedriver

import (
	"os"
	"path/filepath"
)

//go:generate counterfeiter -o ../volmanfakes/fake_file_system.go . FileSystem

// Interface on file system calls in order to facilitate testing
type FileSystem interface {
	MkdirAll(string, os.FileMode) error
	TempDir() string
	Stat(string) (os.FileInfo, error)
	RemoveAll(string) error

	// filepath package
	Abs(path string) (string, error)
}

type realFileSystem struct{}

func NewRealFileSystem() realFileSystem {
	return realFileSystem{}
}

func (f *realFileSystem) MkdirAll(path string, perm os.FileMode) error {
	err := os.MkdirAll(path, perm)
	if err != nil {
		return err
	}

	return os.Chmod(path, perm)

}

func (f *realFileSystem) TempDir() string {
	return os.TempDir()
}

func (f *realFileSystem) Stat(path string) (fi os.FileInfo, err error) {
	return os.Stat(path)
}

func (f *realFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f *realFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}
