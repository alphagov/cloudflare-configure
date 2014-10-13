package main

import (
	"reflect"
)

type ConfigItems map[string]interface{}

func UnionConfigItems(first, second ConfigItems) ConfigItems {
	config := ConfigItems{}

	for key, val := range first {
		config[key] = val
	}

	for key, val := range second {
		config[key] = val
	}

	return config
}

func DifferenceConfigItems(from, to ConfigItems) ConfigItems {
	config := ConfigItems{}

	for key, val := range to {
		if innerVal, _ := from[key]; !reflect.DeepEqual(val, innerVal) {
			config[key] = val
		}
	}

	return config
}
