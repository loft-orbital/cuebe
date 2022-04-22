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
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/loft-orbital/cuebe/cmd/cuebe/prompt"
	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/build"
	"github.com/loft-orbital/cuebe/pkg/instance"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/spf13/cobra"
)

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "apply",
		Aliases:    []string{"deploy"},
		SuggestFor: []string{"install"},
		Short:      "Apply context to k8s cluster.",
		Long: `Apply context to k8s cluster.

Apply uses server-side apply patch to apply the context.
For more information about server-side apply see:
  https://kubernetes.io/docs/reference/using-api/server-side-apply/

It applies every manifests found in the provided context,
grouping them by instance if necessary.
`,
		Example: `
# Apply current directory with an encrypted file override
cuebe apply . main.enc.yaml

# Extract Kubernetes context from <Build>.path.to.context
# You will have to escape the $ sign otherwise you shell environment will try to interpret it
cuebe apply -c \$path.to.context .

# Use one of your available kubectl config context
cuebe apply -c colima .

# Perform a dry-run (do not persist changes)
cuebe apply --dry-run .
`,
		Run: runApply,
	}

	factory.MetaOptionsAware(cmd)
	factory.BuildAware(cmd)
	factory.BuildContextAware(cmd)

	f := cmd.Flags()
	f.StringP("cluster", "c", "", "Kube config context. If starting with a $, it will be extracted from the Build at this CUE path.")
	return cmd
}

func runApply(cmd *cobra.Command, args []string) {
	mfs, build, err := manifetsFrom(cmd)
	cobra.CheckErr(err)

	// group by Instances
	instances := instance.Split(mfs)

	// get kube config
	ctx, err := cmd.Flags().GetString("cluster")
	if strings.HasPrefix(ctx, "$") {
		path := cue.ParsePath(strings.TrimLeft(ctx, "$"))
		cobra.CheckErr(path.Err())
		ctx, err = build.LookupPath(path).String()
		cobra.CheckErr(err)
	}
	if ctx == "" && !prompt.YesNo("Deploy on current kube config context?", cmd.InOrStdin(), cmd.OutOrStdout()) {
		cobra.CheckErr("Canceled by user")
	}
	konfig, err := getK8sConfig(ctx)
	cobra.CheckErr(err)

	// apply changes
	for _, i := range instances {
		cobra.CheckErr(i.Commit(cmd.Context(), konfig, factory.GetMetaOptions(cmd)))
	}
}

func getK8sConfig(context string) (*utils.K8sConfig, error) {
	rconfig, err := utils.DefaultConfig(context)
	if err != nil {
		return nil, fmt.Errorf("could not get config for '%s': %w", context, err)
	}
	konfig, err := utils.NewK8sConfig(rconfig)
	if err != nil {
		return nil, fmt.Errorf("could not get cluster configuration: %w", err)
	}

	return konfig, nil
}

// TODO move that in its own package
func manifetsFrom(cmd *cobra.Command) ([]manifest.Manifest, cue.Value, error) {
	opts := factory.GetBuildOpt(cmd)

	// build
	v, err := build.Build(factory.GetBuildContext(cmd), &load.Config{
		Tags:    opts.Tags,
		TagVars: load.DefaultTagVars(),
	})
	if err != nil {
		return nil, cue.Value{}, fmt.Errorf("could not build context: %w", err)
	}

	// parse paths
	paths := make([]cue.Path, 0, len(opts.Expressions))
	for _, e := range opts.Expressions {
		p := cue.ParsePath(e)
		if p.Err() != nil {
			return nil, v, fmt.Errorf("failed to parse expression %s: %w", e, p.Err())
		}
		paths = append(paths, p)
	}

	// extract manifests
	mfs, err := manifest.Extract(v, paths...)
	if err != nil {
		return nil, v, fmt.Errorf("failed to extract manifests: %w", err)
	}
	if len(mfs) <= 0 {
		return nil, v, fmt.Errorf("no manifest found")
	}

	return mfs, v, nil
}
