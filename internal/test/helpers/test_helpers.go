package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	assert.NoError(t, err)
}
