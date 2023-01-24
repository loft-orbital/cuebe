/*
Copyright © 2021 Loft Orbital

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package mod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/loft-orbital/cuebe/pkg/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
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

	// Check if the module version ref is a tag or a branch
	// will set gco.ReferenceName accordingly
	// TODO: lot of patches and workaround here. We need to clean that up
	if semver.IsValid(mod.Version) || semver.IsValid("v"+mod.Version) {
		// If mod.Version is considered a valid semver, we will presume the module version is a tag
		gco.ReferenceName = plumbing.NewTagReferenceName(mod.Version)
	} else {
		// If the module version is not a valid semver, we will consider it a branch
		gco.ReferenceName = plumbing.NewBranchReferenceName(mod.Version)
	}
	gco.SingleBranch = true

	// set credentials
	if meta.Credentials != nil {
		gco.Auth = &http.BasicAuth{
			Username: meta.Credentials.User,
			Password: meta.Credentials.Token,
		}
	}

	mfs := memfs.New()
	// clone repo
	if _, err := gogit.Clone(memory.NewStorage(), mfs, gco); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	if sd := strings.TrimPrefix(mod.Path, meta.RootPath); sd != "" {
		mfs, err = mfs.Chroot(sd)
		if err != nil {
			return fmt.Errorf("failed to chroot subpath %s %w", sd, err)
		}
	}

	return BillyCopy(fs, mfs)
}
