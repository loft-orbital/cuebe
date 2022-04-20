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
package cmd

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"

	"github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newPackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pack",
		Short: "Package a context.",
		Long: `
Package a context into a tar.gz archive.
This archive can be used later as a base context.

Package respects the .cuebeignore directives.
`,
		Example: `
# Pack current directory
cuebe pack .

# Merge dir1/ and dir2/ and pack them
cuebe pack dir1/ dir2/
`,
		Run: runPackage,
	}

	factory.BuildContextAware(cmd)

	fs := cmd.Flags()
	fs.StringP("output", "o", "cube.tar.gz", "Output file.")

	return cmd
}

func runPackage(cmd *cobra.Command, args []string) {
	ctx := factory.GetBuildContext(cmd)
	fs := ctx.GetAferoFS()
	filename, err := cmd.Flags().GetString("output")
	cobra.CheckErr(err)

	out, err := os.Create(filename)
	cobra.CheckErr(err)
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	cobra.CheckErr(afero.Walk(fs, "", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == "" {
			return nil
		}

		// create header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = path
		//write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// open file
		file, err := fs.Open(path)
		if err != nil {
			return err
		}
		// copy file
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}

		return nil
	}))
}
