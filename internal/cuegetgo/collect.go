package cuegetgo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// MergeImports merges all imports from inputs into a new *ast.File.
func MergeImports(inputs []*ast.File) *ast.File {
	fset := token.NewFileSet()
	f := &ast.File{Package: 0, Name: ast.NewIdent("cuegetgo")}

	for _, in := range inputs {
		for _, imp := range in.Imports {
			name := ""
			if imp.Name != nil {
				name = imp.Name.Name
			}
			astutil.AddNamedImport(fset, f, name, strings.Trim(imp.Path.Value, "\""))
		}
	}

	return f
}

// CollectFiles collects and parses all cuegetgo.go files recursively in root.
func CollectFiles(root string) ([]*ast.File, error) {
	files := []*ast.File{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasSuffix(path, "cue.mod/gen") {
			return fs.SkipDir
		} else if d.Name() == "cuegetgo.go" {
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			files = append(files, f)
		}

		return nil
	})

	return files, err
}
