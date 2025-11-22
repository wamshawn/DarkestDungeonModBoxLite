package files

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CreateTempDir(pattern string) (dir *TempDir, err error) {
	s, createErr := os.MkdirTemp("", pattern)
	if createErr != nil {
		err = createErr
		return
	}
	dir = &TempDir{
		root: s,
		sys:  os.DirFS(s),
	}
	return
}

type TempDir struct {
	root string
	sys  fs.FS
}

func (tf *TempDir) IsDir() bool {
	info, err := fs.Stat(tf.sys, ".")
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (tf *TempDir) Path() string {
	return tf.root
}

func (tf *TempDir) Remove() (err error) {
	err = os.RemoveAll(tf.root)
	return
}

func (tf *TempDir) Sub(name string) (sub *TempDir, err error) {
	dir, subErr := fs.Sub(tf.sys, name)
	if subErr != nil {
		err = subErr
		return
	}
	sub = &TempDir{
		root: filepath.Join(tf.root, name),
		sys:  dir,
	}
	return
}

func (tf *TempDir) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(tf.sys, name)
}

func (tf *TempDir) CreateDir(name string) (*TempDir, error) {
	s := filepath.Join(tf.root, name)
	if err := Mkdir(s); err != nil {
		return nil, err
	}
	return &TempDir{
		root: s,
		sys:  os.DirFS(s),
	}, nil
}

func (tf *TempDir) RemoveDir(name string) error {
	s := filepath.Join(tf.root, name)
	return os.RemoveAll(s)
}

func (tf *TempDir) OpenFile(name string) (*os.File, error) {
	return os.Open(filepath.Join(tf.root, name))
}

func (tf *TempDir) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(tf.sys, name)
}

func (tf *TempDir) ReadFileFull(name string, b []byte) (int, error) {
	s := filepath.Join(tf.root, name)
	file, openErr := os.Open(s)
	if openErr != nil {
		return 0, openErr
	}
	defer file.Close()
	return io.ReadFull(file, b)
}

func (tf *TempDir) WriteFile(name string, b []byte) error {
	s := filepath.Join(tf.root, name)
	return os.WriteFile(s, b, 0644)
}

func (tf *TempDir) Copy(name string, src io.Reader) error {
	s := filepath.Join(tf.root, name)
	dst, dstErr := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if dstErr != nil {
		return dstErr
	}
	defer dst.Close()
	_, cpErr := io.Copy(dst, src)
	return cpErr
}
