package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_int32Abs(t *testing.T) {
	assert.Equal(t, int32(1), int32Abs(-1))
	assert.Equal(t, int32(2), int32Abs(2))
	assert.Equal(t, int32(0), int32Abs(0))
	assert.Equal(t, int32(32), int32Abs(-32))
}
