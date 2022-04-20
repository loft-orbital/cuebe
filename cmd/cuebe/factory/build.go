package factory

import (
	"context"

	"github.com/spf13/cobra"
)

type BuildOpt struct {
	// Expressions are the expressions to extract manifest from.
	Expressions []string
	// Tags are a list of key value used as CUE tags.
	Tags []string
}

type buildKey struct{}

// GetBuildOpt returns the BuildOpt of a command.
// It panics if it hasn't been parsed previously.
// You must use BuildAware to make sure your command is build aware.
func GetBuildOpt(cmd *cobra.Command) *BuildOpt {
	v := cmd.Context().Value(buildKey{})
	return v.(*BuildOpt)
}

// BuildAware marks a command as aware of build options.
// It adds the required flags and PreRunE function to the command.
func BuildAware(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringSliceP("expression", "e", []string{}, "Expressions to extract manifests from. Default to root.")
	f.StringArrayP("tag", "t", []string{}, "Inject boolean or key=value tag.")

	AppendPreRun(cmd, buildPreRun)
}

func buildPreRun(cmd *cobra.Command, args []string) {
	var err error
	fs := cmd.Flags()
	bo := new(BuildOpt)

	bo.Expressions, err = fs.GetStringSlice("expression")
	cobra.CheckErr(err)
	// for some reason setting a string slice with empty string as default in the flag does not work.
	if len(bo.Expressions) <= 0 {
		bo.Expressions = []string{""}
	}

	bo.Tags, err = fs.GetStringArray("tag")
	cobra.CheckErr(err)

	cmd.SetContext(context.WithValue(cmd.Context(), buildKey{}, bo))
}
