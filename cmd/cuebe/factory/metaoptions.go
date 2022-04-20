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
package factory

import (
	"context"

	"github.com/loft-orbital/cuebe/internal/utils"

	"github.com/spf13/cobra"
)

type moKey struct{}

// GetMetaOptions returns the CommonMetaOptions of a command.
// It panics if it's not set
// You must use MetaOptionsAware to make sure your command is properly configured.
func GetMetaOptions(cmd *cobra.Command) utils.CommonMetaOptions {
	v := cmd.Context().Value(moKey{})
	return v.(utils.CommonMetaOptions)
}

// MetaOptionsAware marks a command as aware of a meta options.
// It adds proper flags and the required PreRun function.
func MetaOptionsAware(cmd *cobra.Command) {
	f := cmd.Flags()
	f.BoolP("dry-run", "", false, "Submit server-side request without persisting the resource.")
	f.BoolP("force", "f", false, "Force apply.")
	f.StringP("manager", "m", "cuebe", "Field manager. Override at your own risk.")

	AppendPreRun(cmd, moPreRun)
}

func moPreRun(cmd *cobra.Command, args []string) {
	mo := utils.CommonMetaOptions{}

	f := cmd.Flags()

	force, err := f.GetBool("force")
	cobra.CheckErr(err)
	mo.Force = &force

	dryrun, err := f.GetBool("dry-run")
	cobra.CheckErr(err)
	if dryrun {
		mo.DryRun = []string{"All"}
	}

	mo.FieldManager, err = f.GetString("manager")
	cobra.CheckErr(err)

	cmd.SetContext(context.WithValue(cmd.Context(), moKey{}, mo))
}
