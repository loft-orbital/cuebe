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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/pkg/encoding/yaml"
	"github.com/loft-orbital/cuebe/internal/kubernetes"
	"github.com/spf13/cobra"
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/decrypt"
)

type applyOpts struct {
	Context     string
	EntryPoints []string
	InjectFiles []string
	Dir         string
}

func newApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply cue configuration to kubernetes",
		Long: `
TODO
`,
		Run: applyCmd,
	}

	f := cmd.Flags()
	f.StringP("context", "c", "", "Kubernetes context, or a CUE path to extract it from.")
	f.StringSliceP("inject", "i", []string{}, "Raw YAML files to inject. Can be encrypted with sops.")
	f.StringP("path", "p", "", "Path to load CUE from. Default to current directory")
	return cmd
}

func applyCmd(cmd *cobra.Command, args []string) {
	opts, err := applyParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(applyRun(opts))
}

func applyParse(cmd *cobra.Command, args []string) (*applyOpts, error) {
	opts := &applyOpts{}

	// Context
	c, err := cmd.Flags().GetString("context")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.Context = c

	// InjectFiles
	i, err := cmd.Flags().GetStringSlice("inject")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	opts.InjectFiles = i

	// Dir
	p, err := cmd.Flags().GetString("path")
	if err != nil {
		return nil, fmt.Errorf("Failed parsing args: %w", err)
	}
	if fileInfo, err := os.Stat(p); p != "" && (os.IsNotExist(err) || !fileInfo.IsDir()) {
		return nil, fmt.Errorf("%s does not exist or is not a directory", p)
	}
	opts.Dir = p

	opts.EntryPoints = args
	return opts, nil
}

func applyRun(opts *applyOpts) error {
	// Get root value
	ctx := cuecontext.New()
	bis := load.Instances(opts.EntryPoints, &load.Config{Dir: opts.Dir})
	if len(bis) <= 0 {
		return errors.New("Failed to find any instance")
	}
	if len(bis) > 1 {
		fmt.Println("Found multiple instances, using only the first one")
	}
	bi := bis[0]
	if bi.Err != nil {
		return fmt.Errorf("Error during load: %w", bi.Err)
	}
	value := ctx.BuildInstance(bi)
	if value.Err() != nil {
		return fmt.Errorf("Failed to build instance: %w", value.Err())
	}
	// Merge with raw values
	var err error
	value, err = injectRaw(ctx, value, opts.InjectFiles)
	if err != nil {
		return fmt.Errorf("Failed to inject values: %w", err)
	}

	// Build release
	r, err := kubernetes.NewReleaseFor(value, opts.Context, opts.Context)
	if err != nil {
		return fmt.Errorf("Failed to buid release: %w", err)
	}
	// Deploy Release
	fmt.Printf("Deploying to %s...\n", r.Host())
	return r.Deploy(context.Background())
}

// TODO supports other format (json)
func injectRaw(ctx *cue.Context, v cue.Value, files []string) (cue.Value, error) {
	for _, f := range files {
		b, err := decrypt.File(f, "yaml")
		if err != nil {
			if !errors.Is(err, sops.MetadataNotFound) {
				return v, fmt.Errorf("decryption of %s failed: %w", f, err)
			}
			fmt.Println(f, "does not seem to be encrypted, using plain value")
			b, err = ioutil.ReadFile(f)
			if err != nil {
				return v, fmt.Errorf("could nor read %s: %w", f, err)
			}
		}
		x, err := yaml.Unmarshal(b)
		if err != nil {
			return v, fmt.Errorf("failed to unmarshal %s: %w", f, err)
		}
		w := ctx.BuildExpr(x)
		if w.Err() != nil {
			return v, fmt.Errorf("failed to build %s: %w", f, err)
		}
		v = v.Unify(w)
	}

	return v, v.Err()
}
