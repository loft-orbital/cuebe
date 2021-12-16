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
	"errors"
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/pkg/release"
	"github.com/spf13/cobra"
)

type applyOpts struct {
	Context     string
	EntryPoints []string
	InjectFiles []string
	Expression  string
	Tags        []string
	Dir         string
	DryRun      bool
	Force       bool
}

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "apply",
		Aliases:    []string{"deploy"},
		SuggestFor: []string{"install"},
		Short:      "Apply release to Kubernetes",
		Long: `Apply CUE release to Kubernetes.

Apply uses server-side apply patch to apply the release.
For more information about server-side apply see:
  https://kubernetes.io/docs/reference/using-api/server-side-apply/
`,
		Example: `
# Apply current directory with an encrypted file override
cuebe apply -i main.enc.yaml

# Extract Kubernetes context from CUE path
cuebe apply -c path.to.context

# Perform a dry-run (do not persist changes)
cuebe apply --dry-run
`,
		Run: applyCmd,
	}

	f := cmd.Flags()
	f.StringP("context", "c", "", "Kubernetes context, or a CUE path to extract it from.")
	f.StringSliceP("inject", "i", []string{}, "Inject files into the release. Multiple format supported. Decrypt content with Mozilla sops if extension is .enc.*")
	f.StringP("expression", "e", "", "Expression to extract manifests from. Extract all manifests by default.")
	f.StringArrayP("tag", "t", []string{}, "Inject boolean or key=value tag.")
	f.StringP("path", "p", "", "Path to load CUE from. Default to current directory")
	f.BoolP("dry-run", "", false, "Submit server-side request without persisting the resource.")
	f.BoolP("force", "f", false, "Force apply")
	return cmd
}

func applyCmd(cmd *cobra.Command, args []string) {
	opts, err := applyParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(applyRun(opts))
}

func applyParse(cmd *cobra.Command, args []string) (*applyOpts, error) {
	opts := &applyOpts{}

	// Context
	c, err := cmd.Flags().GetString("context")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Context = c

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

	// DryRun
	dr, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.DryRun = dr

	// DryRun
	f, err := cmd.Flags().GetBool("force")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Force = f

	opts.EntryPoints = args
	return opts, nil
}

func applyRun(opts *applyOpts) error {
	// build config
	cfg := &release.Config{
		Config: &load.Config{
			Dir:     opts.Dir,
			Tags:    opts.Tags,
			TagVars: load.DefaultTagVars(),
		},
		Entrypoints: opts.EntryPoints,
		Orphans:     opts.InjectFiles,
		Context:     context.Background(),
		Target:      cue.ParsePath(opts.Expression),
		KubeContext: opts.Context,
	}

	// load instance
	r, err := release.Load(cfg)
	if err != nil {
		return err
	}

	// ask for confirmation when no k8s context was given
	if r.Context == "" {
		var resp string
		fmt.Print("Deploy on current context? [y/N] ")
		fmt.Scanln(&resp)
		resp = strings.ToLower(strings.TrimSpace(resp))
		if resp != "y" && resp != "yes" {
			// fmt.Println("Canceled by user")
			return errors.New("Canceled by user")
		}
	}

	// apply changes
	po := release.DefaultPatchOptions
	if opts.DryRun {
		po.DryRun = []string{"All"}
	}
	po.Force = &opts.Force
	return r.Apply(cfg.Context, po)
}
