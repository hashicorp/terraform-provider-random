package random_test

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-providers/terraform-provider-random/internal/random"
)

func TestCreateString(t *testing.T) {
	input := random.StringParams{
		MinSpecial:      5,
		Special:         true,
		Length:          10,
		OverrideSpecial: "°",
	}
	result, err := random.CreateString(input)
	stringResult := string(result)

	assert.True(t, utf8.ValidString(stringResult), "a valid string is expected here")
	assert.Equal(t, 10, utf8.RuneCountInString(stringResult))
	require.NoError(t, err)
}
