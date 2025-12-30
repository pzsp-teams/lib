package util_test

import (
	"testing"

	"github.com/pzsp-teams/lib/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestHashWithPepper_KnownVector(t *testing.T) {
	pepper := "pepper123"
	value := "user@example.com"

	const expected = "3cadad0f8d29a1acd80eb42d09b809c554fb7e4bb70051e67193369d56abc021"

	got := util.HashWithPepper(pepper, value)
	assert.Equal(t, expected, got)
}

func TestHashWithPepper_ChangesWhenPepperChanges(t *testing.T) {
	v := "user@example.com"

	h1 := util.HashWithPepper("pepper1", v)
	h2 := util.HashWithPepper("pepper2", v)

	assert.NotEqual(t, h1, h2, "hash should differ when pepper differs")
}

func TestHashWithPepper_ChangesWhenValueChanges(t *testing.T) {
	p := "pepper123"

	h1 := util.HashWithPepper(p, "value1")
	h2 := util.HashWithPepper(p, "value2")

	assert.NotEqual(t, h1, h2, "hash should differ when value differs")
}
