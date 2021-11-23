package mod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/loft-orbital/cuebe/internal/mvs"
	"github.com/loft-orbital/cuebe/pkg/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// Module represents a CUE module
type Module struct {
	root    module.Version
	storage billy.Filesystem
	reqs    mvs.Reqs
}

// New creates a new Module from dir.
// dir has to contain the cue.mod/module.cue file.
func New(dir string) (*Module, error) {
	fs := osfs.New(dir)
	mf, err := modfile.Parse(filepath.Join(fs.Root(), modfile.CueModFile))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("could not parse modfile: %w", err)
		}
		return nil, fmt.Errorf("%s is not a cue module, try running `cue init`", dir)
	}

	return &Module{
		root:    module.Version{Path: mf.Module},
		storage: fs,
		reqs:    modreqs{Root: mf.Module, RootReqs: mf.Require},
	}, nil
}

// Vendor uses MVS algorithm to compute all requirements and vendor them
// in the cue.mod/pkg directory.
func (m *Module) Vendor() error {
	reqs, err := mvs.BuildList(m.root, m.reqs)
	if err != nil {
		return fmt.Errorf("failed to build requirements: %w", err)
	}

	// Vendor requirements
	for _, r := range reqs {
		fs, err := GetFS(r)
		if err != nil {
			return fmt.Errorf("could get %s: %w", r, err)
		}
		dstpath := filepath.Join("cue.mod", "pkg", r.Path)
		dst, err := m.storage.Chroot(dstpath)
		if err != nil {
			return fmt.Errorf("could not chroot to %s: %w", dstpath, err)
		}
		if err := BillyCopy(dst, fs); err != nil {
			return fmt.Errorf("failed to vendor %s: %w", r, err)
		}
		fmt.Printf("vendored %s\n", r)
	}

	return nil
}

type modreqs struct {
	Root     string
	RootReqs []module.Version
}

func (mr modreqs) Required(m module.Version) ([]module.Version, error) {
	if m.Path == mr.Root {
		return mr.RootReqs, nil
	}

	fs, err := GetFS(m)
	if err != nil {
		return nil, fmt.Errorf("failed to get filesystem for %s: %w", m, err)
	}

	reqs, err := GetReqs(fs)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirements for %s: %w", m, err)
	}
	return reqs, err
}

func (mr modreqs) Compare(v, w string) int {
	return semver.Compare(v, w)
}
