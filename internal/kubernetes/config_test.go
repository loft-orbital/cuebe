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
package kubernetes

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
)

func TestExtractContext(t *testing.T) {
	raw := `
context: real: "my-context"
context: fake: 1
`
	ctx := cuecontext.New()
	v := ctx.CompileString(raw)

	// Erroned path
	ktx, err := ExtractContext(v, "[")
	assert.EqualError(t, err, "failed to extract kubernetes context: expected ']', found 'EOF'")
	assert.Empty(t, ktx)

	// Path not found
	ktx, err = ExtractContext(v, "nocontext")
	assert.EqualError(t, err, "failed to extract kubernetes context: field \"nocontext\" not found")
	assert.Empty(t, ktx)

	// Not a string path
	ktx, err = ExtractContext(v, "context.fake")
	assert.EqualError(t, err, "failed to extract kubernetes context: context.fake: cannot use value 1 (type int) as string")
	assert.Empty(t, ktx)

	ktx, err = ExtractContext(v, "context.real")
	assert.NoError(t, err)
	assert.Equal(t, "my-context", ktx)
}
