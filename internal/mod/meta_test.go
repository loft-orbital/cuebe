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
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMeta(t *testing.T) {
	tcs := map[string]struct {
		content  string
		expected *Meta
		err      error
	}{
		"too short": {"path git", nil, fmt.Errorf("Unexpected go-import length")},
		"too long":  {"path git path.git stuff", nil, fmt.Errorf("Unexpected go-import length")},
		"ok":        {"path git path.git", &Meta{RootPath: "path", VCS: "git", RepoURL: "path.git", Credentials: nil}, nil},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual, err := parseMeta(tc.content)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestExtract(t *testing.T) {
	tcs := map[string]struct {
		content  string
		expected *Meta
		err      error
	}{
		"ok":      {`<html><head><meta name="go-import" content="host.com/owner/project git https://host.com/owner/project.git"/></head><body>go get https://host.com/owner/project</body></html>`, &Meta{RootPath: "host.com/owner/project", VCS: "git", RepoURL: "https://host.com/owner/project.git", Credentials: nil}, nil},
		"no meta": {`<html><body>go get https://host.com/owner/project</body></html>`, nil, fmt.Errorf("Could not find go metadata")},
		"eof":     {"", nil, io.EOF},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			actual, err := extract(strings.NewReader(tc.content))
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
