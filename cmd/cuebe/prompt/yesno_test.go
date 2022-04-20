package prompt_test

import (
	. "github.com/loft-orbital/cuebe/cmd/cuebe/prompt"

	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYesNo(t *testing.T) {
	tc := map[string]bool{
		"yes": true,
		"YeS": true,
		"y":   true,
		"Y":   true,
		"yup": false,
		"no":  false,
	}

	for name, expected := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, expected, YesNo("", strings.NewReader(name), ioutil.Discard))
		})
	}
}
