package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyTransforms(t *testing.T) {
	t.Run("applies transforms in declared order to string fields", func(t *testing.T) {
		in := struct {
			Name  string `transform:"lowercase"`
			Code  string `transform:"uppercase"`
			Plain string
		}{Name: "MixedCase", Code: "MixedCase", Plain: "MixedCase"}

		require.NoError(t, ApplyTransforms(&in))

		assert.Equal(t, "mixedcase", in.Name)
		assert.Equal(t, "MIXEDCASE", in.Code)
		assert.Equal(t, "MixedCase", in.Plain)
	})

	t.Run("applies a comma-separated list in order", func(t *testing.T) {
		in := struct {
			LowerWins string `transform:"uppercase,lowercase"`
			UpperWins string `transform:"lowercase,uppercase"`
		}{LowerWins: "MixedCase", UpperWins: "MixedCase"}

		require.NoError(t, ApplyTransforms(&in))

		assert.Equal(t, "mixedcase", in.LowerWins)
		assert.Equal(t, "MIXEDCASE", in.UpperWins)
	})

	t.Run("rejects an unknown transform", func(t *testing.T) {
		in := struct {
			Name string `transform:"titlecase"`
		}{}

		err := ApplyTransforms(&in)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "transform titlecase")
	})

	t.Run("rejects a transform tag on a non-string field", func(t *testing.T) {
		in := struct {
			Count int `transform:"lowercase"`
		}{}

		err := ApplyTransforms(&in)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires a string field")
	})
}
