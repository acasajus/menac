package ostore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FSStore struct {
	rootPath string
}

const (
	FS_ROOT_PATH = "root_path"
)

func (fs *FSStore) Initialize(c map[string]string) error {
	fs.rootPath = c[FS_ROOT_PATH]
	if !filepath.IsAbs(fs.rootPath) {
		return fmt.Errorf("%s is not an absolute path", fs.rootPath)
	}
	fi, err := os.Stat(fs.rootPath)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", fs.rootPath)
	}
	return nil
}

func (fs *FSStore) getPath(so *StoredObject) string {
	return filepath.Join(fs.rootPath, fmt.Sprintf("%s/%s/%s/%s", so.Type, so.Hash[:2], so.Hash[2:4], so.Hash))
}

func (fs *FSStore) Put(so *StoredObject, data io.Reader, length int64) error {
	path := fs.getPath(so)
	_, err := os.Stat(path)
	if err == nil {
		return ErrAlreadyExists
	}
	if err.(*os.PathError).Err.Error() != "no such file or directory" {
		return err
	}
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	hr := NewHashReader(data)
	written, err := io.Copy(f, hr)
	f.Close()
	if err != nil {
		os.Remove(path)
		return err
	}
	if written != length {
		os.Remove(path)
		return fmt.Errorf("Reader length (%d) and defined length differ (%d)", written, length)
	}
	if hr.HexDigest() != so.Hash {
		os.Remove(path)
		return ErrHashMismatch
	}
	return nil
}

func (fs *FSStore) Get(so *StoredObject, data io.Writer) error {
	path := fs.getPath(so)
	f, err := os.Open(path)
	if err != nil {
		if err.(*os.PathError).Err.Error() == "no such file or directory" {
			return ErrNotExists
		}
		return err
	}
	_, err = io.Copy(data, f)
	return err
}

func (fs *FSStore) Delete(so *StoredObject) error {
	err := os.Remove(fs.getPath(so))
	if err != nil && err.(*os.PathError).Err.Error() == "no such file or directory" {
		return ErrNotExists
	}
	return err
}