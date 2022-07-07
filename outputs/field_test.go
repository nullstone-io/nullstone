package outputs

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestGetFields(t *testing.T) {
	type DependencyOutputs struct {
		Output2 string `ns:"output2"`
	}
	type Outputs struct {
		Output1        string `ns:"output1"`
		OptionalOutput string `ns:"optional_output,optional"`

		Dependency DependencyOutputs `ns:",connectionType:some-dependency"`
		Another    DependencyOutputs `ns:",connectionContract:cluster/aws/ecs:ec2"`
	}
	outputsType := reflect.TypeOf(Outputs{})

	want := []Field{
		{
			Field:          outputsType.Field(0),
			Tag:            "output1",
			Name:           "output1",
			ConnectionType: "",
			ConnectionName: "",
			Optional:       false,
		},
		{
			Field:          outputsType.Field(1),
			Tag:            "optional_output,optional",
			Name:           "optional_output",
			ConnectionType: "",
			ConnectionName: "",
			Optional:       true,
		},
		{
			Field:          outputsType.Field(2),
			Tag:            ",connectionType:some-dependency",
			Name:           "",
			ConnectionType: "some-dependency",
			ConnectionName: "",
			Optional:       false,
		},
		{
			Field:              outputsType.Field(3),
			Tag:                ",connectionContract:cluster/aws/ecs:ec2",
			Name:               "",
			ConnectionName:     "",
			ConnectionType:     "",
			ConnectionContract: "cluster/aws/ecs:ec2",
			Optional:           false,
		},
	}

	got := GetFields(outputsType)
	assert.Equal(t, want, got)
}
