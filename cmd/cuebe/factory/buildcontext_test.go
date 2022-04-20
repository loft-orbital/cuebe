package factory

import (
	"context"
	"testing"

	buildctx "github.com/loft-orbital/cuebe/pkg/context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetBuildContext(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// test panic
	assert.Panics(t, func() {
		GetBuildContext(cmd)
	})

	// test nominal
	ctx := buildctx.New()
	cmd.SetContext(context.WithValue(cmd.Context(), bctxKey{}, ctx))
	assert.Equal(t, ctx, GetBuildContext(cmd))
}

func TestBuildContextAware(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	BuildContextAware(cmd)
	assert.NotNil(t, cmd.PreRun)
	assert.NotNil(t, cmd.Args)
}
