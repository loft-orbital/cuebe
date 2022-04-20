package factory

import (
	"context"
	"testing"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetMetaOptions(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	// test panic
	assert.Panics(t, func() {
		GetMetaOptions(cmd)
	})

	// test nominal
	opts := utils.CommonMetaOptions{
		FieldManager: "foo",
	}
	cmd.SetContext(context.WithValue(cmd.Context(), moKey{}, opts))
	assert.Equal(t, opts, GetMetaOptions(cmd))
}

func TestMetaOptionsAware(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	MetaOptionsAware(cmd)
	assert.NotNil(t, cmd.PreRun)
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("manager"))
}
