package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparePathList(t *testing.T) {
	assert.True(t, comparePathList([]string{}, []string{}))

	assert.False(t, comparePathList([]string{"/test1", "/test2"}, []string{"/test1"}))

	assert.True(t, comparePathList([]string{"/test1", "/test2"}, []string{"/test1", "/test2"}))

	assert.False(t, comparePathList([]string{}, []string{"/test1", "/test2"}))
	assert.False(t, comparePathList([]string{"/test1", "/test2"}, []string{}))

	assert.False(t, comparePathList([]string{"/test1", "/test2"}, []string{"/test1", "/testX"}))
}

func TestCompareStringList(t *testing.T) {
	assert.True(t, compareStringList([]string{}, []string{}))

	assert.False(t, compareStringList([]string{"a", "b", "c"}, []string{"a", "b"}))

	assert.True(t, compareStringList([]string{"a", "b", "c"}, []string{"a", "b", "c"}))

	assert.False(t, compareStringList([]string{}, []string{"a", "b", "c"}))
	assert.False(t, compareStringList([]string{"a", "b", "c"}, []string{}))

	assert.False(t, compareStringList([]string{"a", "b", "c"}, []string{"a", "b", "x"}))
}

func TestCombineMaps(t *testing.T) {

	combined := combineMaps(map[string]string{}, map[string]string{})
	assert.Equal(t, 0, len(combined))

	combined = combineMaps(map[string]string{
		"a": "1",
		"b": "2",
	}, map[string]string{})

	assert.Equal(t, 2, len(combined))
	assert.Equal(t, "1", combined["a"])
	assert.Equal(t, "2", combined["b"])

	combined = combineMaps(map[string]string{
		"a": "1",
		"b": "2",
	}, map[string]string{
		"c": "3",
		"d": "4",
	})

	assert.Equal(t, 4, len(combined))
	assert.Equal(t, "1", combined["a"])
	assert.Equal(t, "2", combined["b"])
	assert.Equal(t, "3", combined["c"])
	assert.Equal(t, "4", combined["d"])
}

func TestContains(t *testing.T) {
	assert.False(t, Contains([]string{}, "test"))

	assert.True(t, Contains([]string{"test1", "test2"}, "test1"))
	assert.True(t, Contains([]string{"test1", "test2"}, "test2"))
	assert.False(t, Contains([]string{"test1", "test2"}, "test"))
}
