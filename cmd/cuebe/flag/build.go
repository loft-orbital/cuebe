package flag

import (
	"fmt"

	"github.com/spf13/pflag"
)

const (
	expression      = "expression"
	expressionShort = "e"
	expressionDesc  = "Expressions to extract manifests from. Default to root"

	tag      = "tag"
	tagShort = "t"
	tagDesc  = "Inject boolean or key=value tag."
)

// BuildOpt represents the build options.
type BuildOpt struct {
	// Expressions are the expressions to extract manifest from.
	Expressions []string
	// Tags are a list of key value used as CUE tags.
	Tags []string
}

// AddBuild adds the necessary flags for a build aware command.
func AddBuild(fs *pflag.FlagSet) {
	fs.StringSliceP(expression, expressionShort, []string{}, expressionDesc)
	fs.StringArrayP(tag, tagShort, []string{}, tagDesc)
}

// GetBuild retrieves the build options from flags
func GetBuild(fs *pflag.FlagSet) (*BuildOpt, error) {
	opts := new(BuildOpt)

	// expression
	e, err := fs.GetStringSlice(expression)
	if err != nil {
		return opts, fmt.Errorf("could not parse expressions: %w", err)
	}
	if len(e) <= 0 {
		e = append(e, "")
	}
	opts.Expressions = e

	// tags
	t, err := fs.GetStringArray(tag)
	if err != nil {
		return opts, fmt.Errorf("could not parse tags: %w", err)
	}
	opts.Tags = t

	return opts, nil
}
