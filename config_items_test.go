package main

import (
	"reflect"
	"testing"
)

func TestUnionOfTwoConfigSets(t *testing.T) {
	union := UnionConfigItems(
		ConfigItems{
			"always_online": "off",
			"browser_check": "off",
		},
		ConfigItems{
			"always_online":     "on",
			"browser_cache_ttl": 14400,
		})

	expectedUnion := ConfigItems{
		"always_online":     "on",
		"browser_check":     "off",
		"browser_cache_ttl": 14400,
	}
	if !reflect.DeepEqual(union, expectedUnion) {
		t.Fatal("Expected to only receive the config items that are different", union)
	}
}

func TestDifferenceOfConfigSets(t *testing.T) {
	difference := DifferenceConfigItems(
		ConfigItems{
			"always_online": "off",
			"browser_check": "off",
		},
		ConfigItems{
			"always_online":     "on",
			"browser_check":     "off",
			"browser_cache_ttl": 14400,
		})

	expectedDifference := ConfigItems{
		"always_online":     "on",
		"browser_cache_ttl": 14400,
	}
	if !reflect.DeepEqual(difference, expectedDifference) {
		t.Fatal("Didn't receive the difference we were expecting", difference)
	}
}
