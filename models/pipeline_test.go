package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelinesFilterByType(t *testing.T) {
	pipelines := Pipelines{
		{Name: "alpha", Type: "a"},
		{Name: "beta", Type: "b"},
		{Name: "charlie", Type: "a"},
	}

	pipelines.FilterByType("a")
	expected := Pipelines{
		{Name: "alpha", Type: "a"},
		{Name: "charlie", Type: "a"},
	}

	assert.Equal(t, expected, pipelines)
}

func TestPipelineGet(t *testing.T) {
	pipelines := Pipelines{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "charlie"},
	}

	cases := map[string]bool{
		"alpha":    true,
		"ALPHA":    true,
		"beta":     true,
		"BETA":     true,
		"alhpa":    false,
		"charlie ": false,
	}

	for name, expected := range cases {
		if _, ok := pipelines.Get(name); ok != expected {
			t.Errorf("%s: got %v, expected %v", name, ok, expected)
		}
	}
}

func TestPipelinesDelete(t *testing.T) {
	pipelines := Pipelines{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "charlie"},
	}

	assert.True(t, pipelines.Delete("alpha"))
	assert.True(t, pipelines.Delete("BETA"))

	expected := Pipelines{
		{Name: "charlie"},
	}

	assert.Equal(t, expected, pipelines)
}

func TestPipelinesDeleteError(t *testing.T) {
	pipelines := Pipelines{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "charlie"},
	}

	assert.False(t, pipelines.Delete("delta"))
}

func TestPipelineSort(t *testing.T) {
	pipelines := Pipelines{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "echo"},
		{Name: "beta"},
		{Name: "delta"},
	}

	expected := Pipelines{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "charlie"},
		{Name: "delta"},
		{Name: "echo"},
	}

	pipelines.Sort(true)
	assert.Equal(t, expected, pipelines)

	expected = Pipelines{
		{Name: "echo"},
		{Name: "delta"},
		{Name: "charlie"},
		{Name: "beta"},
		{Name: "alpha"},
	}

	pipelines.Sort(false)
	assert.Equal(t, expected, pipelines)
}
