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
	"github.com/go-git/go-git/v5"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
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
	// if cache doesn't exist
	if _, err := fs.Stat(""); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unexpected stat error: %w", err)
		}
		// module is not cached yet
		if err := Download(mod, fs); err != nil {
			return nil, fmt.Errorf("failed to download module: %w", err)
		}
	} else if !semver.IsValid(mod.Version) && !semver.IsValid("v"+mod.Version) && mod.Version != "latest" {
		// if branch exits on cache update it
		meta, err := GetMeta(mod)
		if err != nil {
			return nil, fmt.Errorf("failed to get meta: %w", err)
		}
		gco := &gogit.CloneOptions{
			URL: meta.RepoURL,
		}
		// set credentials
		if meta.Credentials != nil {
			gco.Auth = &http.BasicAuth{
				Username: meta.Credentials.User,
				Password: meta.Credentials.Token,
			}
		}
		var r *git.Repository
		FetchLatest(r, gco, fs)
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
		URL: meta.RepoURL,
	}
	mfs := memfs.New()

	// set credentials
	if meta.Credentials != nil {
		gco.Auth = &http.BasicAuth{
			Username: meta.Credentials.User,
			Password: meta.Credentials.Token,
		}
	}
	// Check if the module version ref is a tag or a branch
	// will set gco.ReferenceName accordingly
	// TODO: lot of patches and workaround here. We need to clean that up
	if semver.IsValid(mod.Version) || semver.IsValid("v"+mod.Version) {
		// If mod.Version is considered a valid semver, we will presume the module version is a tag
		gco.ReferenceName = plumbing.NewTagReferenceName(mod.Version)
	} else {
		gco.ReferenceName = plumbing.NewBranchReferenceName("main")

	}

	// clone repo
	storage := memory.NewStorage()
	if _, err := gogit.Clone(storage, mfs, gco); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// In case we cloned the main branch, we need to checkout to the commit Hash
	if gco.ReferenceName.Short() == "main" {
		r, err := gogit.Open(storage, mfs)
		if err != nil {
			return fmt.Errorf("failed to open repo: %w", err)
		}
		w, err := r.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree for repo: %w", err)
		}
		w.Checkout(&gogit.CheckoutOptions{Hash: plumbing.NewHash(mod.Version), Force: true})
	}

	if sd := strings.TrimPrefix(mod.Path, meta.RootPath); sd != "" {
		mfs, err = mfs.Chroot(sd)
		if err != nil {
			return fmt.Errorf("failed to chroot subpath %s %w", sd, err)
		}
	}

	return BillyCopy(fs, mfs)
}

// Return latest Tag
func GetLatestTag(gco *gogit.CloneOptions) (string, error) {
	mypath := strings.ReplaceAll(gco.URL, "https:/", "")
	fs, err := CacheLoad(module.Version{Path: mypath})

	var r *git.Repository
	// Fetch and pull for latest tags if repo already exists
	if _, err := fs.Stat(""); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("unexpected stat error: %w", err)
		}
		r, err = gogit.PlainClone(fs.Root(), false, gco)
		if err != nil {
			return "", fmt.Errorf("failed to plain clone repo: %w", err)
		}
	} else {
		r, err = FetchLatest(r, gco, fs)
		if err != nil {
			return "", fmt.Errorf("failed to fetch latest references: %w", err)
		}
	}
	// getting latest tag based on commit timestamp
	var latestTagCommit *object.Commit
	tags, _ := r.Tags()
	var latestTagName string
	err = tags.ForEach(func(t *plumbing.Reference) error {
		revision := plumbing.Revision(t.Name().String())
		tagCommitHash, err := r.ResolveRevision(revision)
		commit, err := r.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}

		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = t.Name().Short()
			t.Type()
		}

		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = t.Name().Short()
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return latestTagName, nil
}

func FetchLatest(r *git.Repository, gco *gogit.CloneOptions, fs billy.Filesystem) (*git.Repository, error) {
	r, err := gogit.PlainOpenWithOptions(fs.Root(), &gogit.PlainOpenOptions{DetectDotGit: false})
	if err != nil {
		return nil, fmt.Errorf("opening module from cache: %w", err)
	}
	// Fetching the repo for latest tags
	err = r.Fetch(&gogit.FetchOptions{
		Auth:       gco.Auth,
		Force:      false,
		Depth:      1,
		Tags:       git.AllTags,
		RemoteName: "origin",
	})
	// New references fetched
	// NoErrAlreadyUpToDate returned as error whenever there are no new updates to fetch
	if err == nil {
		w, err := r.Worktree()
		if err != nil {
			return nil, fmt.Errorf("error on worktree %w", err)
		}
		err = w.Pull(&gogit.PullOptions{
			Auth:  gco.Auth,
			Force: true,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("error on pulling %w", err)
		}
	} else if err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("error on fetching %w", err)
	}
	return r, nil
}

func IsRemoteBranch(repoUrl string, auth transport.AuthMethod, branchName string) (bool, error) {
	// equivalent of git ls-remote
	// Not possible to add flags
	rem := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoUrl},
	})
	// List returned is not sorted
	refs, err := rem.List(&gogit.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return false, fmt.Errorf("error creating remote reference: %w", err)
	}
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			if ref.Name().Short() == branchName {
				return true, nil
			}
		}
	}
	return false, fmt.Errorf("nothing found")
}

// Get the latest tag remotely, without downloading the repo
func GetLatestTagRemote(gco *gogit.CloneOptions) (string, error) {
	// equivalent of git ls-remote
	// Not possible to add flags, get all tags instead and sort them based on semver versionning
	rem := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{gco.URL},
		Fetch: []config.RefSpec{"+refs/tags/*:refs/tags/*"},
	})
	// List returned is not sorted
	refs, err := rem.List(&gogit.ListOptions{
		Auth: gco.Auth,
	})
	if err != nil {
		return "", fmt.Errorf("error listing remote tags: %w", err)
	}
	tags := semver.ByVersion{}
	for _, ref := range refs {
		// It returns also /ref/branch_name, do not take that into account
		// Get only semver ones
		if !ref.Name().IsBranch() && semver.IsValid("v"+ref.Name().Short()) {
			tags = append(tags, ref.Name().Short())
		}
	}
	semver.Sort(tags)
	// Return last sorted tag
	return tags[len(tags)-1], nil
}

// Get the latest commit sha remotely, without downloading the repo
func GetLatestCommitRemote(gco *gogit.CloneOptions, mod module.Version) (string, error) {
	// equivalent of git ls-remote
	// Not possible to add flags, get all tags instead and sort them based on semver versionning
	rem := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{gco.URL},
		Fetch: []config.RefSpec{"+refs/tags/*:refs/tags/*"},
	})
	// List returned is not sorted
	refs, err := rem.List(&gogit.ListOptions{
		Auth: gco.Auth,
	})
	if err != nil {
		return "", fmt.Errorf("error listing remote branches: %w", err)
	}

	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == mod.Version {
			hash := ref.Hash()
			return hash.String(), nil
		}
	}

	return "", fmt.Errorf("error branch %s not found", mod.Version)
}
