package injector

import (
	"errors"
	"fmt"
	"io/fs"

	"cuelang.org/go/cue"
)

// Inject fill a cue value following the injection attributes.
// c.f. https://github.com/loft-orbital/cuebe#inject
func Inject(v cue.Value, fsys fs.FS) cue.Value {
	injections := []Injector{}
	v.Walk(func(v cue.Value) bool {
		// Check for inject
		if a := v.Attribute("inject"); a.Err() == nil {
			injections = append(injections, addInjector(&a, v.Path(), fsys))
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

func addInjector(attr *cue.Attribute, dst cue.Path, fsys fs.FS) Injector {
	t, found, err := attr.Lookup(0, "type")
	if err != nil {
		return NewError(err, dst)
	}
	if !found {
		return NewError(errors.New("missing injector type"), dst)
	}
	switch t {
	case "file":
		return addFileInjector(attr, dst, fsys)
	default:
		return NewError(fmt.Errorf("unsupported injector type %s", t), dst)
	}
}

func addFileInjector(attr *cue.Attribute, dst cue.Path, fsys fs.FS) Injector {
	src, found, err := attr.Lookup(0, "src")
	if err != nil {
		return NewError(err, dst)
	}
	if !found {
		return NewError(errors.New("missing src key for file injector"), dst)
	}

	p, _, err := attr.Lookup(0, "path")
	if err != nil {
		return NewError(err, dst)
	}

	return NewFile(src, p, dst, fsys)
}
