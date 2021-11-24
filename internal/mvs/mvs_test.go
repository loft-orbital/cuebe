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
package mvs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// Example from https://research.swtch.com/vgo-mvs
var vgo = map[module.Version][]module.Version{
	{Path: "A", Version: "v1"}: {{Path: "B", Version: "v1.2"}, {Path: "C", Version: "v1.2"}},

	{Path: "B", Version: "v1.1"}: {{Path: "D", Version: "v1.1"}},
	{Path: "B", Version: "v1.2"}: {{Path: "D", Version: "v1.3"}},

	{Path: "C", Version: "v1.1"}: {},
	{Path: "C", Version: "v1.2"}: {{Path: "D", Version: "v1.4"}},
	{Path: "C", Version: "v1.3"}: {{Path: "F", Version: "v1.1"}},

	{Path: "D", Version: "v1.1"}: {{Path: "E", Version: "v1.1"}},
	{Path: "D", Version: "v1.2"}: {{Path: "E", Version: "v1.1"}},
	{Path: "D", Version: "v1.3"}: {{Path: "E", Version: "v1.2"}},
	{Path: "D", Version: "v1.4"}: {{Path: "E", Version: "v1.2"}},

	{Path: "E", Version: "v1.1"}: {},
	{Path: "E", Version: "v1.2"}: {},
	{Path: "E", Version: "v1.3"}: {},

	{Path: "F", Version: "v1.1"}: {{Path: "G", Version: "v1.1"}},

	{Path: "G", Version: "v1.1"}: {{Path: "F", Version: "v1.1"}},
}

type MockReqs struct {
	Graph map[module.Version][]module.Version
}

func (mrg MockReqs) Required(m module.Version) ([]module.Version, error) {
	return mrg.Graph[m], nil
}

func (mrg MockReqs) Compare(v, w string) int {
	return semver.Compare(v, w)
}

func TestBuildList(t *testing.T) {
	mrg := MockReqs{vgo}
	bl, err := BuildList(module.Version{Path: "A", Version: "v1"}, mrg)
	assert.NoError(t, err)
	assert.ElementsMatch(t, bl, []module.Version{
		{Path: "B", Version: "v1.2"},
		{Path: "C", Version: "v1.2"},
		{Path: "D", Version: "v1.4"},
		{Path: "E", Version: "v1.2"},
	})
}
