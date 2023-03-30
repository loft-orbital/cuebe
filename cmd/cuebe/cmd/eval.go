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
	"github.com/spf13/cobra"
	"fmt"
	"cuelang.org/go/cue"
	"github.com/loft-orbital/cuebe/pkg/build"
	"cuelang.org/go/cue/load"
)

func newEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "eval",
		SuggestFor: []string{"render", "template"},
		Short:      "Equivalent of cue eval command.",
		Long: `
Will evaluate the cue definitions without requiring concrete values
the same way 'cue eval' would do.
		`,
		Example: `
# Export current directory
cuebe eval .
`,
		Run: runEval,
	}

	factory.BuildAware(cmd)
	factory.BuildContextAware(cmd)

	return cmd
}

func runEval(cmd *cobra.Command, args []string) {
    // Build opts, shall return a BuildOpt{} struct
    opts := factory.GetBuildOpt(cmd)

    // build the cuebe context, and return a cue.Value with the unified cue structure
    v, err := build.Build(factory.GetBuildContext(cmd), &load.Config{
        Tags:    opts.Tags,
        TagVars: load.DefaultTagVars(),
    })
    if err != nil {
        fmt.Errorf("could not build context: %w", err)
    }

    // Parse paths expressions (-e argument)
    paths := make([]cue.Path, 0, len(opts.Expressions))
    for _, e := range opts.Expressions {
        p := cue.ParsePath(e)
        if p.Err() != nil {
            fmt.Errorf("failed to parse expression %s: %w", e, p.Err())
        }
        paths = append(paths, p)
    }

    // Output the cue evaluation WITHOUT concrete values
    // If we have no paths, we dump the whole unified values
    if len(paths) == 0 {
        fmt.Printf("%v", v)
    } else {
        for _, p := range paths {
            node := v.LookupPath(p)
            fmt.Printf("%v", node)
        }
    }
}
