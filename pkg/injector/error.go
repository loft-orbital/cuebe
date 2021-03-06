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
package injector

import (
	"fmt"

	"cuelang.org/go/cue"
)

// Error represents an injection error.
type Error struct {
	path cue.Path
	err  error
}

// NewError creates a new error injector.
func NewError(err error, dstPath cue.Path) *Error {
	return &Error{path: dstPath, err: err}
}

// Inject injects the error into the target and return the result value.
func (e *Error) Inject(target cue.Value) cue.Value {
	return target.FillPath(e.path, e)
}

func (e *Error) Error() string {
	return fmt.Sprintf("injection error: %v", e.err)
}
