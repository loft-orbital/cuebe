package context

import (
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path"

	"github.com/spf13/afero"
)

type Context struct {
	fs afero.Fs
}

func New() *Context {
	return &Context{fs: afero.NewMemMapFs()}
}

func FromArgs(args []string) (*Context, error) {
	ctx := New()

	for _, arg := range args {
		if !path.IsAbs(arg) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("could not get working directory: %w", err)
			}
			arg = path.Join(cwd, arg)
		}
		if err := ctx.Add(afero.NewBasePathFs(afero.NewOsFs(), arg)); err != nil {
			return nil, fmt.Errorf("could not add %s to context: %w", arg, err)
		}
	}

	return ctx, nil
}

// GetFS returns the standard fs.FS underlying filesystem.
// The returned filesystem is read only.
func (c *Context) GetFS() iofs.FS {
	return afero.NewIOFS(c.GetAferoFS())
}

// GetAferoFS returns the uderlying afero.Fs of this context.
func (c *Context) GetAferoFS() afero.Fs {
	return afero.NewReadOnlyFs(c.fs)
}

// Add copies the content of fs into this Context.
// The content of fs takes priority if there is a conflict.
func (c *Context) Add(fs afero.Fs) error {
	return Copy(c.fs, fs)
}

// Copy copies src afero.Fs into dst.
func Copy(dst, src afero.Fs) error {
	return afero.Walk(src, "", func(path string, info iofs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == "" {
			return nil
		}

		switch info.Mode() & os.ModeType {
		case iofs.ModeDir:
			if err := dst.MkdirAll(path, info.Mode().Perm()); err != nil {
				return err
			}
		default:
			_, err := CopyFile(dst, src, path, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to copy %s: %w", path, err)
			}
		}

		return nil
	})
}

// CopyFile copy a file from src to dst, keeping the same file name and path.
func CopyFile(dst, src afero.Fs, name string, stat iofs.FileMode) (int64, error) {
	srcF, err := src.Open(name)
	if err != nil {
		return 0, fmt.Errorf("could not open file: %w", err)
	}
	defer srcF.Close()

	dstF, err := dst.OpenFile(name, os.O_RDWR|os.O_CREATE, stat)
	if err != nil {
		return 0, fmt.Errorf("could not create file: %w", err)
	}
	defer dstF.Close()

	return io.Copy(dstF, srcF)
}
