package factory

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetBuild(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// test panic
	assert.Panics(t, func() {
		GetBuildOpt(cmd)
	})

	// test nominal
	opts := BuildOpt{
		Expressions: []string{"foo"},
		Tags:        []string{"bar", "baz"},
	}
	cmd.SetContext(context.WithValue(cmd.Context(), buildKey{}, &opts))
	assert.Equal(t, &opts, GetBuildOpt(cmd))
}

func TestBuildAware(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	BuildAware(cmd)
	assert.NotNil(t, cmd.PreRun)
	assert.NotNil(t, cmd.Flags().Lookup("expression"))
	assert.NotNil(t, cmd.Flags().Lookup("tag"))
}
