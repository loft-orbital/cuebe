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
package mod

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrivate(t *testing.T) {
	tcs := map[string]struct {
		pattern  string
		module   string
		expected bool
	}{
		"empty":  {"", "github.com/proj/mod", false},
		"any":    {"*", "github.com/proj/mod", true},
		"host":   {"github.com", "github.com/proj/mod", true},
		"nohost": {"gitlab.com", "github.com/proj/mod", false},
		"mod":    {"github.com/*/mod", "github.com/proj/mod", true},
		"nomod":  {"github.com/*/nomod", "github.com/proj/mod", false},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			CuePrivatePattern = tc.pattern
			assert.Equal(t, tc.expected, IsPrivate(tc.module))
		})
	}
}

func TestCredentialsFor(t *testing.T) {
	os.Setenv("MY_HOST_COM_TOKEN", "token")
	os.Setenv("MY_HOST_COM_USER", "user")
	defer func() {
		os.Unsetenv("MY_HOST_COM_TOKEN")
		os.Unsetenv("MY_HOST_COM_USER")
	}()

	usr, pwd := CredentialsFor("my.host.com")
	assert.Equal(t, "user", usr)
	assert.Equal(t, "token", pwd)
}
