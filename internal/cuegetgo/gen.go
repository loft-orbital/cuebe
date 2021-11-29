package cuegetgo

import (
	"context"
	"fmt"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/loft-orbital/cuebe/internal/mod"
	"golang.org/x/sync/errgroup"
)

// GenGO generates CUE definitions for all imports in cuegetgo files.
func GenGO(root string) (billy.Filesystem, error) {
	fs := memfs.New()

	// gather imports
	files, err := CollectFiles(root)
	if err != nil {
		return nil, fmt.Errorf("failed to collect cuegetgo.go files: %w", err)
	}
	cgg := MergeImports(files)

	// create a temporary working directory
	wd, err := ioutil.TempDir("", "cuebe-gen")
	if err != nil {
		return nil, fmt.Errorf("failed to create working directory: %w", err)
	}
	defer os.RemoveAll(wd)
	bwd := osfs.New(wd)

	excmd := func(cmd string, args ...string) error {
		c := exec.Command(cmd, args...)
		c.Dir = bwd.Root()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// cue mod init
	if err := excmd("cue", "mod", "init", "github.com/loft-orbital/cuegetgo"); err != nil {
		return nil, fmt.Errorf("failed to run `cue mod init`: %w", err)
	}

	// go mod init
	if err := excmd("go", "mod", "init", "github.com/loft-orbital/gogetgo"); err != nil {
		return nil, fmt.Errorf("failed to run `go mod init`: %w", err)
	}

	// put cuegetgo file
	cggf, err := bwd.Create("cuegetgo.go")
	if err := format.Node(cggf, token.NewFileSet(), cgg); err != nil {
		return nil, fmt.Errorf("failed to create cuegetgo.go file: %w", err)
	}

	// go mod tidy
	if err := excmd("go", "mod", "tidy", "-compat=1.17"); err != nil {
		return nil, fmt.Errorf("failed to run `go mod tidy`: %w", err)
	}

	// go mod vendor
	if err := excmd("go", "mod", "vendor"); err != nil {
		return nil, fmt.Errorf("failed to run `go mod vendor`: %w", err)
	}

	// cue get go
	eg, _ := errgroup.WithContext(context.Background())
	for _, imp := range cgg.Imports {
		pkg := strings.Trim(imp.Path.Value, "\"")
		eg.Go(func() error {
			if err := excmd("cue", "get", "go", pkg); err != nil {
				return fmt.Errorf("%s: %w", pkg, err)
			}
			fmt.Printf("definitions generated for %s\n", pkg)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate cue definitions: %w", err)
	}

	gwd, err := bwd.Chroot(filepath.Join("cue.mod", "gen"))
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to gen folder: %w", err)
	}

	if err := mod.BillyCopy(fs, gwd); err != nil {
		return nil, fmt.Errorf("failed to cache gen folder in memory: %w", err)
	}

	return fs, nil
}
