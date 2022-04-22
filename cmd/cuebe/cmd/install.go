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
	"github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/loft-orbital/cuebe/pkg/instance"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Cuebe to K8s cluster.",
		Long: `
Install Cuebe custom resource definitions to the k8s cluster.
		`,
		Example: `
# Install to current config context.
cuebe install

# Same but targetting my-cluster.
cuebe install -c my-cluster
`,
		Run: runInstall,
	}

	factory.MetaOptionsAware(cmd)

	f := cmd.Flags()
	f.StringP("cluster", "c", "", "Kube config context.")
	return cmd
}

func runInstall(cmd *cobra.Command, args []string) {
	// get kube config
	ctx, err := cmd.Flags().GetString("cluster")
	cobra.CheckErr(err)
	konfig, err := getK8sConfig(ctx)
	cobra.CheckErr(err)

	// install CRDs
	cobra.CheckErr(instance.InstallCRD(cmd.Context(), konfig, factory.GetMetaOptions(cmd)))

	cmd.Println("Installation successful.")
	cmd.Println("Enjoy working with cuebe!")
}
