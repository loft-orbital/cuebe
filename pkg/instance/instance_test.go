package instance

import (
	"testing"

	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	mfs := []manifest.Manifest{
		newUniqueManifest().WithInstance("potato"),
		newUniqueManifest().WithInstance("potato"),
		newUniqueManifest().WithInstance("potato"),

		newUniqueManifest().WithInstance("tomato"),
		newUniqueManifest().WithInstance("tomato"),
		newUniqueManifest().WithInstance("tomato"),
		newUniqueManifest().WithInstance("tomato"),
		newUniqueManifest().WithInstance("tomato"),

		newUniqueManifest(),
		newUniqueManifest(),
	}
	instances := Split(mfs)
	assert.Len(t, instances, 3)
	assert.Len(t, instances[0].Manifests(), 3)
	assert.Equal(t, "potato", instances[0].String())
	assert.Len(t, instances[1].Manifests(), 5)
	assert.Equal(t, "tomato", instances[1].String())
	assert.Len(t, instances[2].Manifests(), 2)
	assert.IsType(t, (*Orphan)(nil), instances[2])
}
