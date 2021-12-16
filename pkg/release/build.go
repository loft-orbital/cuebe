package release

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/loft-orbital/cuebe/pkg/injector"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Inject(v cue.Value) cue.Value {
	injections := []injector.Injector{}
	v.Walk(func(v cue.Value) bool {
		// Check for inject
		if a := v.Attribute("inject"); a.Err() == nil {
			injections = append(injections, addInjector(&a, v.Path()))
			return false // no nested injection
		}
		return true
	}, nil)
	// do inject
	for _, i := range injections {
		v = i.Inject(v)
	}
	return v
}

func Extract(v cue.Value) ([]unstructured.Unstructured, error) {
	objs := []unstructured.Unstructured{}
	var errs error

	v.Walk(func(v cue.Value) bool {
		if a := v.Attribute("ignore"); a.Err() == nil {
			return false
		}

		// Check if could be a manifest
		k, vs := v.LookupPath(cue.MakePath(cue.Str("kind"))), v.LookupPath(cue.MakePath(cue.Str("apiVersion")))
		if k.Kind() == cue.StringKind && vs.Kind() == cue.StringKind {
			o, err := manifest.New(v).ToObj() // TODO parallelize extraction
			if err != nil {
				errs = multierror.Append(errs, err)
			} else {
				objs = append(objs, o)
			}
			return false
		}
		return true
	}, nil)

	return objs, errs
}

func addInjector(attr *cue.Attribute, dst cue.Path) injector.Injector {
	t, found, err := attr.Lookup(0, "type")
	if err != nil {
		return injector.NewError(err, dst)
	}
	if !found {
		return injector.NewError(errors.New("Injector type not found"), dst)
	}
	switch t {
	case "file":
		return addFileInjector(attr, dst)
	default:
		return injector.NewError(fmt.Errorf("Unsupported injector type %s", t), dst)
	}
}

func addFileInjector(attr *cue.Attribute, dst cue.Path) injector.Injector {
	src, found, err := attr.Lookup(0, "src")
	if err != nil {
		return injector.NewError(err, dst)
	}
	if !found {
		return injector.NewError(errors.New("Missing 'src' key for type file"), dst)
	}

	p, found, err := attr.Lookup(0, "path")
	if err != nil {
		return injector.NewError(err, dst)
	}
	if !found {
		p = "$"
	}

	return injector.NewFile(src, p, dst)
}
