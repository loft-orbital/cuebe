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
	"container/list"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"golang.org/x/mod/module"
)

var cache billy.Filesystem

// CacheDir returns the cuebe cache directory (~/.cache/cuebe).
func CacheDir() (billy.Filesystem, error) {
	if cache == nil {
		hd, err := os.UserHomeDir()
		if err != nil {
			hd = "~"
		}
		cd := filepath.Join(hd, ".cache", "cuebe")
		if err := os.MkdirAll(cd, 0755); err != nil {
			return nil, err
		}
		cache = osfs.New(cd)
	}

	return cache, nil
}

// CacheLoad retrieve the module's filesystem from the cache.
// This filesystem can be empty (os.ErrNotExist).
func CacheLoad(m module.Version) (billy.Filesystem, error) {
	cd, err := CacheDir()
	if err != nil {
		return nil, fmt.Errorf("getting cache: %w", err)
	}

	//fmt.Printf("%+v\n", m)
	return cd.Chroot(m.String())
}

// CacheStore store a module filesystem to the global cache.
// It returns the cached filesystem.
func CacheStore(m module.Version, fs billy.Filesystem) (billy.Filesystem, error) {
	cd, err := CacheDir()
	if err != nil {
		return nil, fmt.Errorf("getting cache: %w", err)
	}

	mFS, err := cd.Chroot(m.String())
	if err != nil {
		return nil, fmt.Errorf("getting module cache: %w", err)
	}

	if err := BillyCopy(mFS, fs); err != nil {
		return nil, fmt.Errorf("copying module to cache: %w", err)
	}

	return mFS, nil
}

// BillyCopy copy srcFS into dstFS billy filesystems.
func BillyCopy(dstFS, srcFS billy.Filesystem) error {
	dirs := list.New()
	dirs.PushBack("") // push root dir

	for dirs.Len() > 0 {
		dir := dirs.Back() // get current iteration directory
		// create destination dir
		if err := dstFS.MkdirAll(dir.Value.(string), 0755); err != nil {
			return fmt.Errorf("copying %s -> %s: %w", srcFS.Root(), dstFS.Root(), err)
		}
		// read all files and directories
		files, err := srcFS.ReadDir(dir.Value.(string))
		if err != nil {
			return fmt.Errorf("copying %s -> %s: %w", srcFS.Root(), dstFS.Root(), err)
		}

		for _, file := range files {
			fn := filepath.Join(dir.Value.(string), file.Name())
			if !file.IsDir() {
				if err := billyCopyFile(dstFS, srcFS, fn); err != nil {
					return fmt.Errorf("copying %s -> %s: %w", srcFS.Root(), dstFS.Root(), err)
				}
			} else {
				dirs.PushBack(fn) // push directory to visit later
			}
		}
		dirs.Remove(dir) // we're done with this directory
	}
	return nil
}

func billyCopyFile(dstFS, srcFS billy.Filesystem, filename string) error {
	src, err := srcFS.Open(filename)
	if err != nil {
		return fmt.Errorf("copying %s: %w", filename, err)
	}
	defer src.Close()
	dst, err := dstFS.Create(filename)
	if err != nil {
		return fmt.Errorf("copying %s: %w", filename, err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copying %s: %w", filename, err)
	}
	return nil
}
