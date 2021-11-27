package cli

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func Build(cmd *cobra.Command, out billy.Filesystem, siteroot string) error {
	if err := out.MkdirAll("", 0755); err != nil {
		return fmt.Errorf("could not create command directory: %w", err)
	}

	rel := strings.TrimPrefix(out.Root(), siteroot)
	rel = strings.Trim(rel, "/")

	rnav := &NavNode{Id: cmd.Name(), Child: make(map[string]*NavNode)}
	if err := doc.GenMarkdownTreeCustom(cmd, out.Root(), navBuilder(rnav, siteroot), absLinkHandler(rel)); err != nil {
		return fmt.Errorf("failed to generate markdown tree: %w", err)
	}

	nav, err := out.Create("_sidebar.md")
	if err != nil {
		return fmt.Errorf("failed to create sidebar file: %w", err)
	}
	defer nav.Close()
	fmt.Fprintln(nav, "* [Home](/)")
	rnav.Format(nav, 0)
	return nil
}

func navBuilder(nav *NavNode, root string) (filePrepender func(string) string) {
	filePrepender = func(path string) string {
		fn := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		cs := strings.Split(fn, "_")

		link, err := filepath.Rel(root, path)
		if err != nil {
			log.Fatal(err)
		}
		n := nav.Get(cs...)
		n.Link = link

		return ""
	}

	return
}

func absLinkHandler(root string) func(string) string {
	return func(rel string) string {
		return filepath.Join(root, rel)
	}
}
