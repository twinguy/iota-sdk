package uuidutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Generate two UUIDs and ensure they're different
	id1 := New()
	id2 := New()

	assert.NotEqual(t, id1, id2, "Two generated UUIDs should be different")
	assert.NotEqual(t, id1, uuid.Nil, "Generated UUID should not be nil")
	assert.NotEqual(t, id2, uuid.Nil, "Generated UUID should not be nil")
}

func TestParse(t *testing.T) {
	validUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	id, err := Parse(validUUID)

	assert.NoError(t, err, "Parsing a valid UUID should not return an error")
	assert.Equal(t, validUUID, id.String(), "Parsed UUID should match the original string")

	_, err = Parse("invalid-uuid")
	assert.Error(t, err, "Parsing an invalid UUID should return an error")
}

func TestMustParse(t *testing.T) {
	validUUID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	id := MustParse(validUUID)

	assert.Equal(t, validUUID, id.String(), "Parsed UUID should match the original string")

	// MustParse should panic with invalid input, so we test that with a defer/recover
	defer func() {
		r := recover()
		assert.NotNil(t, r, "MustParse should panic with invalid input")
	}()

	MustParse("invalid-uuid")
}

func TestIsNil(t *testing.T) {
	assert.True(t, IsNil(uuid.Nil), "IsNil should return true for uuid.Nil")
	assert.False(t, IsNil(New()), "IsNil should return false for a newly generated UUID")
}

func TestString(t *testing.T) {
	id := MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	assert.Equal(t, "f47ac10b-58cc-4372-a567-0e02b2c3d479", String(id), "String should return the correct string representation")
}
