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
		"ok":        {"path git path.git", &Meta{"path", "git", "path.git", nil}, nil},
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
		"ok":      {`<html><head><meta name="go-import" content="host.com/owner/project git https://host.com/owner/project.git"/></head><body>go get https://host.com/owner/project</body></html>`, &Meta{"host.com/owner/project", "git", "https://host.com/owner/project.git", nil}, nil},
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
