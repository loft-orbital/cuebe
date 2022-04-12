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
	"errors"
	"fmt"
	"os"
	"path"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/cmd/cuebe/flag"
	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/build"
	bctxt "github.com/loft-orbital/cuebe/pkg/context"
	"github.com/loft-orbital/cuebe/pkg/instance"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type applyOpts struct {
	flag.BuildOpt
	BuildContext *bctxt.Context
	DryRun       bool
	Force        bool
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
	flag.AddBuild(f)
	f.BoolP("dry-run", "", false, "Submit server-side request without persisting the resource.")
	f.BoolP("force", "f", false, "Force apply")
	return cmd
}

func applyCmd(cmd *cobra.Command, args []string) {
	opts, err := applyParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(applyRun(cmd, opts))
}

func applyParse(cmd *cobra.Command, args []string) (*applyOpts, error) {
	opts := new(applyOpts)

	bopts, err := flag.GetBuild(cmd.Flags())
	if err != nil {
		return opts, fmt.Errorf("could not get build options: %w", err)
	}
	opts.BuildOpt = *bopts

	// dry-run
	dr, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.DryRun = dr

	// force
	f, err := cmd.Flags().GetBool("force")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Force = f

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

func applyRun(cmd *cobra.Command, opts *applyOpts) error {
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
		return errors.New("could not find any manifests")
	}

	// group by Instances
	instances := instance.Split(mfs)

	// get kube config
	// TODO retrieve from args + build
	rconfig, err := utils.DefaultConfig("")
	if err != nil {
		return fmt.Errorf("could not get default k8s config: %w", err)
	}
	konfig, err := utils.NewK8sConfig(rconfig)
	if err != nil {
		return fmt.Errorf("could not get cluster configuration: %w", err)
	}

	// apply changes
	po := utils.CommonMetaOptions{
		FieldManager: "cuebe",
		Force:        &opts.Force,
	}
	if opts.DryRun {
		po.DryRun = []string{"All"}
	}
	for _, i := range instances {
		if err := i.Commit(cmd.Context(), konfig, po); err != nil {
			return fmt.Errorf("could not commit instance %s: %w", i, err)
		}
	}

	return nil
}
