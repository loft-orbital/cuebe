/*
Copyright © 2021 Loft Orbital

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
