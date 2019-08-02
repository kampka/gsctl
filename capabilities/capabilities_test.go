package capabilities

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testInput struct {
	Provider       string
	ReleaseVersion string
}

var capabilityTests = []struct {
	in  testInput
	out []CapabilityDefinition
}{
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "1.2.3",
		},
		[]CapabilityDefinition{},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "6.1.2",
		},
		[]CapabilityDefinition{AvailabilityZones},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "6.4.0",
		},
		[]CapabilityDefinition{Autoscaling, AvailabilityZones},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "9.0.0",
		},
		[]CapabilityDefinition{Autoscaling, AvailabilityZones, NodePools},
	},
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "9.1.2",
		},
		[]CapabilityDefinition{Autoscaling, AvailabilityZones, NodePools},
	},
	{
		testInput{
			Provider:       "kvm",
			ReleaseVersion: "9.1.2",
		},
		[]CapabilityDefinition{},
	},
}

var failingCapabilityTests = []struct {
	in  testInput
	out string
}{
	{
		testInput{
			Provider:       "aws",
			ReleaseVersion: "1.2.3.4",
		},
		"Invalid Semantic Version",
	},
}

func TestGetCapabilities(t *testing.T) {
	for index, tt := range capabilityTests {
		cap, err := GetCapabilities(tt.in.Provider, tt.in.ReleaseVersion)
		if err != nil {
			t.Errorf("Test %d: Error: %s", index, err)
		}
		if !cmp.Equal(cap, tt.out) {
			t.Errorf("Test %d: Expected %#v but got %#v", index, tt.out, cap)
		}
	}
}

func TestFailingCapabilities(t *testing.T) {
	for index, tt := range failingCapabilityTests {
		_, err := GetCapabilities(tt.in.Provider, tt.in.ReleaseVersion)
		if err == nil {
			t.Errorf("Test %d: Expected error, got nil", index)
		}
		if err.Error() != tt.out {
			t.Errorf("Test %d: Expected error '%s', got '%s'", index, tt.out, err.Error())
		}
	}
}