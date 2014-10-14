package main

import (
	"reflect"
)

type ConfigMismatch struct {
	Missing ConfigItems
}

func (c ConfigMismatch) Error() string {
	return "Config found that is present in the CDN config but not in the local config"
}

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
