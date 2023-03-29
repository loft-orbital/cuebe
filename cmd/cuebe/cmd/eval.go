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
// 	"cuelang.org/go/cue"
	"github.com/loft-orbital/cuebe/pkg/build"
	"cuelang.org/go/cue/load"
// 	"cuelang.org/go/cue/cuecontext"
//     "cuelang.org/go/cue/format"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newEvalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "eval",
		SuggestFor: []string{"render", "template"},
		Short:      "Equivalent of cue eval command.",
		Long: `
Will evaluate the cue definition without requiring concrete values and produce a JSON output
the same as 'cue eval' would do.
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
// 	opts := factory.GetBuildOpt(cmd)
//
// 	// build
// 	v, err := build.Build(factory.GetBuildContext(cmd), &load.Config{
// 		Tags:    opts.Tags,
// 		TagVars: load.DefaultTagVars(),
// 	})
// 	if err != nil {
// 		return nil, cue.Value{}, fmt.Errorf("could not build context: %w", err)
// 	}
//
// 	// parse paths
// 	paths := make([]cue.Path, 0, len(opts.Expressions))
// 	for _, e := range opts.Expressions {
// 		p := cue.ParsePath(e)
// 		fmt.Printf("%q", p)
// 		if p.Err() != nil {
// 			return nil, v, fmt.Errorf("failed to parse expression %s: %w", e, p.Err())
// 		}
// 		paths = append(paths, p)
// 	}

    evalFrom(cmd)
}

func evalFrom(cmd *cobra.Command) { //([]manifest.Manifest, cue.Value, error) {
    opts := factory.GetBuildOpt(cmd)

    // build the cuebe context
    v, err := build.Build(factory.GetBuildContext(cmd), &load.Config{
        Tags:    opts.Tags,
        TagVars: load.DefaultTagVars(),
    })
    if err != nil {
        fmt.Errorf("could not build context: %w", err)
    }

    // print the value
    fmt.Printf("// %%v\n%v\n\n// %%# v\n%# v\n", v, v)
    fmt.Println("End of pouet")
}
