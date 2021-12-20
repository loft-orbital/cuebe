package release

import (
	"context"
	"fmt"
	"sort"

	"cuelang.org/go/cue"
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

type Config struct {
	*load.Config

	Entrypoints []string
	Orphans     []string
	Context     context.Context
	KubeContext string
	Target      cue.Path
}

func Load(cfg *Config) (*Release, error) {
	// load root instance
	u, err := unifier.Load(cfg.Entrypoints, cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to load instance: %w", err)
	}

	// add oprhan files
	oeg, _ := errgroup.WithContext(cfg.Context)
	for _, f := range cfg.Orphans {
		oeg.Go(func() error { return u.AddFile(f) })
	}
	if err := oeg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to add orphans: %w", err)
	}
	v := u.Unify()
	if v.Err() != nil {
		return nil, fmt.Errorf("failed to build release: %w", v.Err())
	}

	// injection
	v = Inject(v)

	// check for error
	if err := v.Validate(); err != nil {
		return nil, err
	}

	// get k8s context
	ktx, err := extractContext(v, cfg.KubeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kubernetes context: %w", err)
	}

	// extract manifests
	objs, err := Extract(v.LookupPath(cfg.Target))
	if err != nil {
		return nil, fmt.Errorf("failed to extract manifests: %w", err)
	}
	sort.Sort(SortableUnstructured(objs))

	return &Release{
		Context: ktx,
		Objects: objs,
	}, nil
}

func extractContext(v cue.Value, c string) (string, error) {
	if c == "" {
		return c, nil
	}

	p := cue.ParsePath(c)
	if p.Err() != nil {
		return c, nil
	}

	vktx := v.LookupPath(p)
	if vktx.Err() != nil {
		if vktx.Err().Error() == fmt.Sprintf("field \"%s\" not found", c) {
			return c, nil
		}
		return "", fmt.Errorf("unexpected error: %w", vktx.Err())
	}
	sktx, err := vktx.String()
	if err != nil {
		return "", fmt.Errorf("context cannot be resolve as a string: %w", err)
	}
	return sktx, nil
}