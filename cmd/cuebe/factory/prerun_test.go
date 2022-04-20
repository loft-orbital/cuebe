package factory_test

import (
	"testing"

	. "github.com/loft-orbital/cuebe/cmd/cuebe/factory"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendPreRun(t *testing.T) {
	run := make([]string, 0, 2)
	cmd := &cobra.Command{}

	AppendPreRun(cmd, func(cmd *cobra.Command, args []string) {
		run = append(run, "first")
	})
	AppendPreRun(cmd, func(cmd *cobra.Command, args []string) {
		run = append(run, "second")
	})

	require.NotNil(t, cmd.PreRun)
	cmd.PreRun(nil, nil)

	assert.Len(t, run, 2)
	assert.Equal(t, []string{"first", "second"}, run)
}
