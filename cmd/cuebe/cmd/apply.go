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

	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/internal/kubernetes"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"github.com/spf13/cobra"
)

type applyOpts struct {
	Context     string
	EntryPoints []string
	InjectFiles []string
	Expressions []string
	Tags        []string
	Dir         string
	DryRun      bool
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
	f.StringArrayP("expression", "e", []string{}, "Expressions to extract manifests from. Extract all manifests by default.")
	f.StringArrayP("tag", "t", []string{}, "Inject boolean or key=value tag.")
	f.StringP("path", "p", "", "Path to load CUE from. Default to current directory")
	f.BoolP("dry-run", "", false, "Submit server-side request without persisting the resource.")
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
	e, err := cmd.Flags().GetStringArray("expression")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Expressions = e

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

	opts.EntryPoints = args
	return opts, nil
}

func applyRun(opts *applyOpts) error {
	// load instance
	u, err := unifier.Load(opts.EntryPoints, &load.Config{
		Dir:  opts.Dir,
		Tags: opts.Tags,
	})
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
	r, err := kubernetes.NewReleaseFor(u.Unify(), opts.Expressions, opts.Context, opts.Context)
	if err != nil {
		return fmt.Errorf("Failed to buid release: %w", err)
	}
	// deploy Release
	fmt.Printf("Deploying to %s...\n", r.Host())
	return r.Deploy(context.Background(), opts.DryRun)
}
