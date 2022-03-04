package context

import (
	"fmt"
	"io"
	"io/fs"
	iofs "io/fs"

	"github.com/spf13/afero"
)

type Context struct {
	fs afero.Fs
}

func New() *Context {
	return &Context{fs: afero.NewMemMapFs()}
}

// GetFS returns the standard fs.FS underlying filesystem.
// The returned filesystem is read only.
func (c *Context) GetFS() iofs.FS {
	return afero.NewIOFS(afero.NewReadOnlyFs(c.fs))
}

// Add copies the content of fs into this Context.
// The content of fs takes priority if there is a conflict.
func (c *Context) Add(fs afero.Fs) error {
	return Copy(c.fs, fs)
}

// Copy copies src afero.Fs into dst.
func Copy(dst, src afero.Fs) error {
	return afero.Walk(src, "", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch info.Mode() {
		case iofs.ModeDir:
			if err := dst.MkdirAll(path, info.Mode().Perm()); err != nil {
				return err
			}
		default:
			_, err := CopyFile(dst, src, path)
			if err != nil {
				return fmt.Errorf("failed to copy %s: %w", path, err)
			}
		}

		return nil
	})
}

// CopyFile copy a file from src to dst, keeping the same file name and path.
func CopyFile(dst, src afero.Fs, name string) (int64, error) {
	srcF, err := src.Open(name)
	if err != nil {
		return 0, fmt.Errorf("could not open file: %w", err)
	}
	defer srcF.Close()

	dstF, err := dst.Create(name)
	if err != nil {
		return 0, fmt.Errorf("could not create file: %w", err)
	}
	defer dstF.Close()

	return io.Copy(dstF, srcF)
}
