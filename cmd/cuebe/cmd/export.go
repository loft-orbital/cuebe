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
	"fmt"
	"os"

	"github.com/loft-orbital/cuebe/internal/kubernetes"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"github.com/spf13/cobra"
)

type exportOpts struct {
	EntryPoints []string
	InjectFiles []string
	Expressions []string
	Dir         string
}

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "export",
		SuggestFor: []string{"render", "template"},
		Short:      "Export manifests as YAML.",
		Long:       "Export CUE release as a kubectl-compatible multi document YAML manifest.",
		Example: `
# Export current directory with an encrypted file override
cuebe export -i main.enc.yaml
`,
		Run: exportCmd,
	}

	f := cmd.Flags()
	f.StringSliceP("inject", "i", []string{}, "Raw YAML files to inject. Can be encrypted with sops.")
	f.StringArrayP("expression", "e", []string{}, "Expressions to extract manifests from. Extract all manifests by default.")
	f.StringP("path", "p", "", "Path to load CUE from. Default to current directory")
	return cmd
}

func exportCmd(cmd *cobra.Command, args []string) {
	opts, err := exportParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(exportRun(cmd, opts))
}

func exportParse(cmd *cobra.Command, args []string) (*exportOpts, error) {
	opts := &exportOpts{}

	// InjectFiles
	i, err := cmd.Flags().GetStringSlice("inject")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.InjectFiles = i

	// Expression
	e, err := cmd.Flags().GetStringArray("expression")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Expressions = e

	// Dir
	p, err := cmd.Flags().GetString("path")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	if fileInfo, err := os.Stat(p); p != "" && (os.IsNotExist(err) || !fileInfo.IsDir()) {
		return nil, fmt.Errorf("%s does not exist or is not a directory", p)
	}
	opts.Dir = p

	opts.EntryPoints = args
	return opts, nil
}

func exportRun(cmd *cobra.Command, opts *exportOpts) error {
	// load instance
	u, err := unifier.Load(opts.EntryPoints, opts.Dir)
	if err != nil {
		return fmt.Errorf("Failed to load instance: %w", err)
	}
	// inject orphan files
	for _, f := range opts.InjectFiles {
		if err := u.AddFile(f); err != nil {
			return fmt.Errorf("failed to inject %s: %w", f, err)
		}
	}

	// build release
	r, err := kubernetes.NewReleaseFor(u.Unify(), opts.Expressions, "", "")
	if err != nil {
		return fmt.Errorf("Failed to buid release: %w", err)
	}
	return r.Render(cmd.OutOrStdout())
}
