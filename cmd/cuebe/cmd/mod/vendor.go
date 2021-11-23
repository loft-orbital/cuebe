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

	"github.com/loft-orbital/cuebe/internal/mod"
	"github.com/spf13/cobra"
)

type vendorOpts struct {
	ModRoot string
}

func newVendorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vendor",
		Short: "vendor requirements in cue.mod/pkg",
		Args:  cobra.MaximumNArgs(1),
		Run:   vendorCmd,
	}

	return cmd
}

func vendorCmd(cmd *cobra.Command, args []string) {
	opts, err := vendorParse(cmd, args)
	cobra.CheckErr(err)
	cobra.CheckErr(vendorRun(opts))
}

func vendorParse(cmd *cobra.Command, args []string) (*vendorOpts, error) {
	opts := &vendorOpts{}
	if len(args) > 0 {
		opts.ModRoot = args[0]
	}
	return opts, nil
}

func vendorRun(opts *vendorOpts) error {
	m, err := mod.New(opts.ModRoot)
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}

	return m.Vendor()
}
