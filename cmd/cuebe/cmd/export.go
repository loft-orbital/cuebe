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
	"context"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/pkg/release"
	"github.com/spf13/cobra"
)

type exportOpts struct {
	EntryPoints []string
	InjectFiles []string
	Expression  string
	Tags        []string
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
	f.StringP("expression", "e", "", "Expressions to extract manifests from. Extract all manifests by default.")
	f.StringArrayP("tag", "t", []string{}, "Inject boolean or key=value tag.")
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
	e, err := cmd.Flags().GetString("expression")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Expression = e

	// Tags
	t, err := cmd.Flags().GetStringArray("tag")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Tags = t

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
	r, err := release.Load(&release.Config{
		Config: &load.Config{
			Dir:     opts.Dir,
			Tags:    opts.Tags,
			TagVars: load.DefaultTagVars(),
		},
		Entrypoints: opts.EntryPoints,
		Orphans:     opts.InjectFiles,
		Context:     context.Background(),
		Target:      cue.ParsePath(opts.Expression),
	})
	if err != nil {
		return err
	}

	return r.Render(cmd.OutOrStdout())
}
