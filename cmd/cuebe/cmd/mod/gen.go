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
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/loft-orbital/cuebe/internal/cuegetgo"
	"github.com/loft-orbital/cuebe/internal/mod"
	"github.com/spf13/cobra"
)

type genOpts struct {
	ModRoot string
}

func newGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "gen requirements in cue.mod/pkg",
		Args:  cobra.MaximumNArgs(1),
		Run:   genCmd,
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
	mfs, err := cuegetgo.GenGO(opts.ModRoot)
	if err != nil {
		return fmt.Errorf("failed to build definitions: %w", err)
	}
	gen := osfs.New(filepath.Join(opts.ModRoot, "cue.mod", "gen"))
	if err := os.RemoveAll(gen.Root()); err != nil {
		return fmt.Errorf("failed to clean gen directory: %w", err)
	}
	return mod.BillyCopy(gen, mfs)
}
