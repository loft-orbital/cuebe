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
package mod

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mod",
	Short: "manage CUE modules.",
	Long: `cuebe mod provides access to operations on CUE modules.
Most mod subcommands use a custom implementation of the MVS algorithm.
This is a temporary solution until [github.com/cue-lang#85](https://github.com/cue-lang/cue/issues/851) is addressed.

#### Adding requirements

Simply add a new entry to the 'require' key in cue.mod/module.cue file:

~~~cue
require: [
  { path: "github.com/tomato/ketchup", version: "v1.0.3" },
]
~~~

#### Private modules

When dealing with private modules, cuebe offers two solutions:
- leveraging credentials within your ~/.netrc file
- exporting **HOST_ADDRESS_TOKEN** and **HOST_ADDRESS_USER** environment variable (e.g. **GITHUB_COM_TOKEN** for github.com)
`,
}

func init() {
	RootCmd.AddCommand(
		newGenCmd(),
		newVendorCmd(),
	)
}
