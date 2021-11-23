package mod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/loft-orbital/cuebe/pkg/modfile"
	"golang.org/x/mod/module"
)

// GetFS returns the cached filesystem for mod.
// If the module is not yet cached, it will be first downloaded and then cached.
func GetFS(mod module.Version) (billy.Filesystem, error) {
	fs, err := CacheLoad(mod)
	if err != nil {
		return nil, fmt.Errorf("loading module from cache: %w", err)
	}

	if _, err := fs.Stat(""); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unexpected stat error: %w", err)
		}
		// module is not cached yet
		if err := Download(mod, fs); err != nil {
			return nil, fmt.Errorf("failed to download module: %w", err)
		}
	}

	return fs, nil
}

// GetReqs returns the dependencies of a CUE module located in fs.
func GetReqs(fs billy.Filesystem) ([]module.Version, error) {
	mf, err := modfile.Parse(filepath.Join(fs.Root(), modfile.CueModFile))
	if err != nil {
		return nil, fmt.Errorf("failed to get modfile: %w", err)
	}
	return mf.Require, nil
}

// Download clone and store repository worktree into fs.
func Download(mod module.Version, fs billy.Filesystem) error {
	meta, err := GetMeta(mod)
	if err != nil {
		return fmt.Errorf("failed to get meta: %w", err)
	}
	gco := &gogit.CloneOptions{
		URL:   meta.RepoURL,
		Depth: 1,
	}

	// set reference
	if mod.Version != "v0.0.0" {
		gco.ReferenceName = plumbing.NewTagReferenceName(mod.Version)
		gco.SingleBranch = true
	}

	// set credentials
	if meta.Credetials != nil {
		gco.Auth = &http.BasicAuth{
			Username: meta.Credetials.User,
			Password: meta.Credetials.Token,
		}
	}

	// clone repo
	if _, err := gogit.Clone(memory.NewStorage(), fs, gco); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	return nil
}
