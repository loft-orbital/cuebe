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
	"strings"

	"cuelang.org/go/cue"
	"github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/loft-orbital/cuebe/cmd/cuebe/prompt"
	"github.com/loft-orbital/cuebe/pkg/instance"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "delete",
		SuggestFor: []string{"remove", "uninstall"},
		Short:      "Delete all instances found in Build.",
		Long: `
Delete all instances found in provided context from the k8s cluster.

It first group manifests found in the context by instance.
Then it deletes those instances.
Cuebe delete respects the deletion policy annotation "instance.cuebe.loftorbital.com/deletion-policy".
		`,
		Example: `
# Delete all instances in the current dir
cuebe delete .

# Same but doing a dry-run
cuebe delete --dry-run .

# Delete using Kubernetes context from <Build>.path.to.context
cuebe apply -c .release.context .

# Delete using one of your available kubectl config context
cuebe apply -c colima .
`,
		Run: runDelete,
	}

	factory.MetaOptionsAware(cmd)
	factory.BuildAware(cmd)
	factory.BuildContextAware(cmd)

	f := cmd.Flags()
	f.StringP("cluster", "c", "", "Kube config context. If starting with a . (dot), it will be extracted from the Build at this CUE path.")
	return cmd
}

func runDelete(cmd *cobra.Command, args []string) {
	mfs, build, err := manifestFrom(cmd)
	cobra.CheckErr(err)

	// group by Instances
	instances := instance.Split(mfs)

	// get kube config
	ctx, err := cmd.Flags().GetString("cluster")
	cobra.CheckErr(err)
	if strings.HasPrefix(ctx, ".") {
		path := cue.ParsePath(strings.TrimLeft(ctx, "."))
		cobra.CheckErr(path.Err())
		ctx, err = build.LookupPath(path).String()
		cobra.CheckErr(err)
	}
	if ctx == "" && !prompt.YesNo("Delete from current kube config context?", cmd.InOrStdin(), cmd.OutOrStdout()) {
		cobra.CheckErr("Canceled by user")
	}
	konfig, err := getK8sConfig(ctx)
	cobra.CheckErr(err)

	// apply changes
	for _, i := range instances {
		cobra.CheckErr(i.Delete(cmd.Context(), konfig, factory.GetMetaOptions(cmd)))
	}
}
