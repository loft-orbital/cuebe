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
package cmd

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/muesli/coral"
)

var (
	version string
	commit  string
	date    string
)

type versionOpts struct{}

func newVersionCmd() *coral.Command {
	cmd := &coral.Command{
		Use:   "version",
		Short: "Print current version",
		Run:   versionCmd,
	}

	return cmd
}

func versionCmd(cmd *coral.Command, args []string) {
	opts, err := versionParse(cmd, args)
	coral.CheckErr(err)
	coral.CheckErr(versionRun(opts))
}

func versionParse(cmd *coral.Command, args []string) (*versionOpts, error) {
	opts := &versionOpts{}
	return opts, nil
}

func versionRun(opts *versionOpts) error {
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 0, '\t', tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\t", "Version:", version)
	fmt.Fprintf(w, "\n%s\t%s\t", "Go version:", runtime.Version())
	fmt.Fprintf(w, "\n%s\t%s\t", "Git commit:", commit)
	fmt.Fprintf(w, "\n%s\t%s\t", "Built:", date)
	fmt.Fprintf(w, "\n%s\t%s/%s\t", "OS/Arch:", runtime.GOOS, runtime.GOARCH)

	return nil
}
