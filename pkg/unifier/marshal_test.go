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
package unifier

import (
	"testing"

	"cuelang.org/go/cue/cuecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshallerFor(t *testing.T) {
	tcs := map[string]struct {
		ext         string
		expected    interface{}
		shouldError bool
	}{
		"cue":         {".cue", &CUEUnmarshaller{}, false},
		"yaml":        {".yaml", &YAMLUnmarshaller{}, false},
		"yml":         {".yml", &YAMLUnmarshaller{}, false},
		"json":        {".json", &JSONUnmarshaller{}, false},
		"unsupported": {".unsupported", nil, true},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			u, err := UnmarshallerFor(tc.ext)
			if tc.shouldError {
				assert.Error(t, err)
				assert.Nil(t, u)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tc.expected, u)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tcs := map[string]struct {
		content     string
		expected    string
		u           Unmarshaller
		shouldError bool
	}{
		"cue nominal":  {"foo: \"bar\"", "{\"foo\":\"bar\"}", &CUEUnmarshaller{}, false},
		"cue error":    {"foo: ", "", &CUEUnmarshaller{}, true},
		"json nominal": {"{\"foo\":\"bar\"}", "{\"foo\":\"bar\"}", &JSONUnmarshaller{}, false},
		"json error":   {"nojson", "", &JSONUnmarshaller{}, true},
		"yaml nominal": {"foo: bar", "{\"foo\":\"bar\"}", &YAMLUnmarshaller{}, false},
		"yaml error":   {"\tnoyaml", "", &YAMLUnmarshaller{}, true},
	}

	ctx := cuecontext.New()
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			v, err := tc.u.Unmarshal([]byte(tc.content), ctx)
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				actual, err := v.MarshalJSON()
				require.NoError(t, err)
				assert.Equal(t, tc.expected, string(actual))
			}
		})
	}
}
