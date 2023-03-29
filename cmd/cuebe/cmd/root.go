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
	"os"
	"time"

	"github.com/loft-orbital/cuebe/cmd/cuebe/cmd/mod"
	"github.com/loft-orbital/cuebe/pkg/log"
	"github.com/spf13/cobra"
)

type cancelKey struct{}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cuebe",
	Short: "Handle CUE kubernetes release",
	Long: `cuebe handles CUE Kubernetes release.

  Find more information at: https://github.com/loft-orbital/cuebe
`,
	DisableAutoGenTag: true,
	PersistentPreRun:  setContext,
	PersistentPostRun: cleanContext,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().Duration("timeout", 2*time.Minute, "Timeout, accpet any valid go Duration.")

	RootCmd.AddCommand(
		newApplyCmd(),
		newDeleteCmd(),
		newExportCmd(),
		newEvalCmd(),
		newInstallCmd(),
		newPackCmd(),
		newVersionCmd(),
		mod.RootCmd,
	)
}

func setContext(cmd *cobra.Command, args []string) {
	timeout, err := cmd.Flags().GetDuration("timeout")
	cobra.CheckErr(err)

	// add timeout
	withTimeout, cancel := context.WithTimeout(cmd.Context(), timeout)
	withCancel := context.WithValue(withTimeout, cancelKey{}, cancel)
	// add logger
	withLog := log.WithLogger(withCancel, log.NewIOLogger(cmd.OutOrStdout(), cmd.ErrOrStderr()))

	cmd.SetContext(withLog)
}

func cleanContext(cmd *cobra.Command, args []string) {
	v := cmd.Context().Value(cancelKey{})
	if cancel, ok := v.(context.CancelFunc); ok {
		// need to release the context
		cancel()
	}
}
