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

	buildctx "github.com/loft-orbital/cuebe/pkg/context"

	"github.com/spf13/cobra"
)

type bctxKey struct{}

// GetBuildContext returns the build Context of a command.
// It panics if it's not set
// You must use BuildContextAware to make sure your command is Context aware.
func GetBuildContext(cmd *cobra.Command) *buildctx.Context {
	v := cmd.Context().Value(bctxKey{})
	return v.(*buildctx.Context)
}

// BuildContextAware marks a command as aware of a build Context.
// It adds an args validation and the required PreRunE function.
func BuildContextAware(cmd *cobra.Command) {
	cmd.Args = cobra.MinimumNArgs(1)
	AppendPreRun(cmd, bctxPreRun)
}

func bctxPreRun(cmd *cobra.Command, args []string) {
	bctx, err := buildctx.FromArgs(args)
	cobra.CheckErr(err)

	cmd.SetContext(context.WithValue(cmd.Context(), bctxKey{}, bctx))
}
