package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFeedback(t *testing.T) {
	assert.Equal(t, "kind/name verb", buildFeedback("Kind", "name", "verb", false))
	assert.Equal(t, "kind/name verb (server dry run)", buildFeedback("Kind", "name", "verb", true))
}
