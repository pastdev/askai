package optional

import (
	"testing"

	"github.com/openai/openai-go/v3/packages/param"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestValueOf(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		v, err := valueOf[bool]("true")
		require.NoError(t, err)
		require.Equal(t, true, v)
	})

	t.Run("float64", func(t *testing.T) {
		v, err := valueOf[float64]("1.9")
		require.NoError(t, err)
		require.Equal(t, 1.9, v)
	})

	t.Run("int64", func(t *testing.T) {
		v, err := valueOf[int64]("7")
		require.NoError(t, err)
		require.Equal(t, int64(7), v)
	})
}

func OptionalVarPTester[T optionalTypes](
	t *testing.T,
	args []string,
	expectedValid bool,
	expected T,
) {
	var optV param.Opt[T]

	cmd := &cobra.Command{
		Run: func(_ *cobra.Command, _ []string) {
			if expectedValid {
				require.True(t, optV.Valid())
				require.Equal(t, expected, optV.Value)
			} else {
				require.False(t, optV.Valid())
			}
		},
	}

	VarP(cmd.Flags(), &optV, "optv", "", "testing")

	cmd.SetArgs(args)

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestOptionalVarP(t *testing.T) {
	t.Run("implicit bool", func(t *testing.T) {
		OptionalVarPTester(t, []string{"--optv"}, true, true)
	})

	t.Run("implicit explicit bool true", func(t *testing.T) {
		OptionalVarPTester(t, []string{"--optv=true"}, true, true)
	})

	t.Run("implicit explicit bool false", func(t *testing.T) {
		OptionalVarPTester(t, []string{"--optv=false"}, true, false)
	})

	t.Run("int", func(t *testing.T) {
		OptionalVarPTester(t, []string{"--optv", "7"}, true, int64(7))
	})

	t.Run("float64", func(t *testing.T) {
		OptionalVarPTester(t, []string{"--optv", "1.0"}, true, 1.0)
	})

	t.Run("dont set flag", func(t *testing.T) {
		OptionalVarPTester(t, []string{}, false, true)
	})
}
