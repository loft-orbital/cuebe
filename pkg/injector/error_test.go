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
	"errors"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	e := NewError(errors.New("the error"), cue.ParsePath("."))
	assert.IsType(t, (*Error)(nil), e)
}

func TestErrorInject(t *testing.T) {
	ctx := cuecontext.New()
	v := ctx.CompileString("foo: _")
	e := NewError(errors.New("the error"), cue.ParsePath("foo"))
	assert.NoError(t, v.Err())

	v = e.Inject(v)
	assert.EqualError(t, v.Err(), "injection error: the error")
}
