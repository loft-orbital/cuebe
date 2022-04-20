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
	"io"

	"github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "export",
		SuggestFor: []string{"render", "template"},
		Short:      "Export manifests as YAML.",
		Long: `
Export CUE release as a kubectl-compatible multi document YAML manifest.
If --output is set, manifests will be written here, one file by instances.
		`,
		Example: `
# Export current directory with an encrypted file override
cuebe export -i main.enc.yaml
`,
		Run: runExport,
	}

	factory.BuildAware(cmd)
	factory.BuildContextAware(cmd)

	return cmd
}

func runExport(cmd *cobra.Command, args []string) {
	mfs, _, err := manifetsFrom(cmd)
	cobra.CheckErr(err)

	// render
	w := cmd.OutOrStdout()
	encoder := yaml.NewEncoder(w)
	defer encoder.Close()
	for _, m := range mfs {
		_, err := io.WriteString(w, "---\n")
		cobra.CheckErr(err)
		cobra.CheckErr(encoder.Encode(m.Object))
	}
}
