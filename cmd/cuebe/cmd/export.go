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
	"io"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/cmd/cuebe/flag"
	"github.com/loft-orbital/cuebe/pkg/build"
	bctxt "github.com/loft-orbital/cuebe/pkg/context"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type exportOpts struct {
	flag.BuildOpt
	BuildContext *bctxt.Context
	OutputDir    string
}

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
		Run: exportCmd,
	}

	f := cmd.Flags()
	flag.AddBuild(f)
	f.StringP("output", "o", "", "Directory to write manifests to. If empty it will print into stdout")
	return cmd
}

func exportCmd(cmd *cobra.Command, args []string) {
	opts, err := exportParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(exportRun(cmd, opts))
}

func exportParse(cmd *cobra.Command, args []string) (*exportOpts, error) {
	opts := new(exportOpts)

	bopts, err := flag.GetBuild(cmd.Flags())
	if err != nil {
		return opts, fmt.Errorf("could not get build options: %w", err)
	}
	opts.BuildOpt = *bopts

	// output
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("could not get output flag: %w", err)
	}
	opts.OutputDir = output

	// context
	opts.BuildContext = bctxt.New()
	for _, arg := range args {
		// TODO move that to a function in the context package
		if !path.IsAbs(arg) {
			cwd, err := os.Getwd()
			if err != nil {
				return opts, fmt.Errorf("could not get working directory: %w", err)
			}
			arg = path.Join(cwd, arg)
		}
		if err := opts.BuildContext.Add(afero.NewBasePathFs(afero.NewOsFs(), arg)); err != nil {
			return opts, fmt.Errorf("could not add %s to context: %w", arg, err)
		}
	}

	return opts, nil
}

func exportRun(cmd *cobra.Command, opts *exportOpts) error {
	// build
	v, err := build.Build(opts.BuildContext, &load.Config{
		Tags:    opts.Tags,
		TagVars: load.DefaultTagVars(),
	})
	if err != nil {
		return fmt.Errorf("could not build context: %w", err)
	}

	// parse paths
	paths := make([]cue.Path, 0, len(opts.Expressions))
	for _, e := range opts.Expressions {
		p := cue.ParsePath(e)
		if p.Err() != nil {
			return fmt.Errorf("could not parse path %s: %w", e, p.Err())
		}
		paths = append(paths, p)
	}

	// extract manifests
	mfs, err := manifest.Extract(v, paths...)
	if err != nil {
		return fmt.Errorf("could not extract manifests: %w", err)
	}
	if len(mfs) <= 0 {
		return fmt.Errorf("could not find any manifests")
	}

	// render
	w := cmd.OutOrStdout()
	encoder := yaml.NewEncoder(w)
	defer encoder.Close()
	for _, m := range mfs {
		if _, err := io.WriteString(w, "---\n"); err != nil {
			return fmt.Errorf("failed to write %s header: %w", m.GetName(), err)
		}
		if err := encoder.Encode(m.Object); err != nil {
			return fmt.Errorf("failed to marshal %s: %w", m.GetName(), err)
		}
	}

	return nil
}
