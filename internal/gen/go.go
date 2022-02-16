package gen

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/loft-orbital/cuebe/pkg/modfile"
)

type void struct{}

var null void

func GenGoPkg(root string) error {
	pkgs, err := CollectGoPkg(root)
	if err != nil {
		return fmt.Errorf("could not collect go pkg: %w", err)
	}

	excmd := func(cmd string, args ...string) error {
		c := exec.Command(cmd, args...)
		c.Dir = root
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if _, err := os.Stat(path.Join(root, "go.mod")); errors.Is(err, os.ErrNotExist) {
		// go.mod file dos not exist, let's create it
		mf, err := modfile.Parse(path.Join(root, modfile.CueModFile))
		if err != nil {
			return fmt.Errorf("cannot get cue mod file: %w", err)
		}
		if err := excmd("go", "mod", "init", mf.Module); err != nil {
			return fmt.Errorf("go mod init failed: %w", err)
		}
	}

	// go get
	args := append([]string{"get", "-d"}, pkgs...)
	if err := excmd("go", args...); err != nil {
		return fmt.Errorf("failed to download packages: %w", err)
	}

	// cue get go
	args = make([]string, 0, len(pkgs)+2)
	args = append(args, "get", "go")
	for _, p := range pkgs {
		args = append(args, strings.SplitN(p, "@", 2)[0])
	}
	if err := excmd("cue", args...); err != nil {
		return fmt.Errorf("failed to generate cue definitions: %w", err)
	}

	return nil
}

func CollectGoPkg(root string) ([]string, error) {
	dedupe := make(map[string]void)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasSuffix(path, "cue.mod/gen") {
			// skip gen dir
			return fs.SkipDir
		} else if d.Type().IsRegular() && strings.HasSuffix(path, modfile.CueModFile) {
			mf, err := modfile.Parse(path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			for _, pkg := range mf.GoDefinitions {
				dedupe[pkg] = null
			}
		}
		return nil
	})

	pkgs := make([]string, 0, len(dedupe))
	for k := range dedupe {
		pkgs = append(pkgs, k)
	}

	return pkgs, err
}
