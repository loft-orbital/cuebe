package modfile

import (
	"fmt"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"
)

func Parse(file string) (*File, error) {
	cf, err := parser.ParseFile(file, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse modfile: %w", err)
	}

	ctx := cuecontext.New()
	v := ctx.BuildFile(cf)
	if v.Err() != nil {
		return nil, fmt.Errorf("could not build modfile: %w", err)
	}
	f := &File{}
	if err := v.Decode(f); err != nil {
		return nil, fmt.Errorf("failed to decode modfile: %w", err)
	}

	return f, nil
}
