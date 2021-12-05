package release

import (
	"context"
	"errors"
	"fmt"

	"cuelang.org/go/cue/load"
	"github.com/loft-orbital/cuebe/pkg/unifier"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Release struct {
	Context string

	// Objects managed by this release
	Objects []unstructured.Unstructured
}

func Load(entrypoints, orphans []string, cfg *load.Config, ctx context.Context) (*Release, error) {
	// Load root instance
	u, err := unifier.Load(entrypoints, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load instance: %w", err)
	}
	// Add oprhan files
	oeg, _ := errgroup.WithContext(ctx)
	for _, f := range orphans {
		oeg.Go(func() error { return u.AddFile(f) })
	}
	if err := oeg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to add orphans: %w", err)
	}
	v := u.Unify()
	if v.Err() != nil {
		return nil, fmt.Errorf("failed to build release: %w", v.Err())
	}

	return nil, errors.New("Not implemented")
}
