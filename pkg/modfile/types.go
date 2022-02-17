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
package modfile

import (
	"path"

	"golang.org/x/mod/module"
)

// CueModFile is the default location of a CUE module file.
var CueModFile = path.Join("cue.mod", "module.cue")

type File struct {
	// Module name, sometime called path.
	Module string `json:"module"`
	// Module requirements.
	Require []module.Version `json:"require,omitempty"`
	// Go definition to generate
	GoDefinitions []string `json:"godef,omitempty"`
}
