package utils

import (
	"github.com/hashicorp/go-multierror"
)

// CollectErrors reduce erros in cerr into a single go error
// using github.com/hashicorp/go-multierror.
func CollectErrors(cerr <-chan error) error {
	var errs error
	for e := range cerr {
		if e != nil {
			errs = multierror.Append(errs, e)
		}
	}

	return errs
}
