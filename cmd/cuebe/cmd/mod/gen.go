/*
Copyright © 2021 loft-orbital

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
package mod

import (
	"os"

	"github.com/loft-orbital/cuebe/internal/gen"
	"github.com/spf13/cobra"
)

type genOpts struct {
	ModRoot string
}

func newGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "genenerates CUE definitions from Go modules.",
		Long: `Collects all cuegetgo.go files (including those in cue.mod/pkg/**) and
generates CUE definitions for all imported packages.

To add a new package to generate definitions for, include it in the godef list of your cue.mod/module.cue file.
You can fix version by appending @version/

~~~cue
module: "github.com/company/module"

godef: [
  "k8s.io/api/apps/v1",
  "k8s.io/api/batch/v1@v0.23.3",
  ]
~~~
`,
		Args: cobra.MaximumNArgs(1),
		Run:  genCmd,
	}

	return cmd
}

func genCmd(cmd *cobra.Command, args []string) {
	opts, err := genParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(genRun(opts))
}

func genParse(cmd *cobra.Command, args []string) (*genOpts, error) {
	opts := &genOpts{}
	if len(args) > 0 {
		opts.ModRoot = args[0]
	} else {
		path, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.ModRoot = path
	}
	return opts, nil
}

func genRun(opts *genOpts) error {
	return gen.GenGoPkg(opts.ModRoot)
}
