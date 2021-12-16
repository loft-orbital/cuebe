package release

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

// Render write a kubectl compatible yaml manifest to w.
func (r *Release) Render(w io.Writer) error {
	for _, obj := range r.Objects {
		by, err := yaml.Marshal(obj.Object)
		if err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
		if _, err := w.Write([]byte("---\n")); err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
		if _, err := w.Write(by); err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
	}
	return nil
}
