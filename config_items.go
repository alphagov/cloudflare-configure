package main

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
)

type ConfigMismatch struct {
	Missing ConfigItems
}

func (c ConfigMismatch) Error() string {
	return "Config found that is present in the CDN config but not in the local config"
}

type ConfigItems map[string]interface{}

type ConfigItemsForUpdate map[string]ConfigItemForUpdate

type ConfigItemForUpdate struct {
	Current  interface{}
	Expected interface{}
}

func CompareConfigItemsForUpdate(current, expected ConfigItems) (ConfigItemsForUpdate, error) {
	union := UnionConfigItems(current, expected)
	differenceCurrentAndUnion := DifferenceConfigItems(current, union)
	differenceExpectedAndUnion := DifferenceConfigItems(expected, union)

	if len(differenceExpectedAndUnion) > len(differenceCurrentAndUnion) {
		return nil, ConfigMismatch{Missing: differenceExpectedAndUnion}
	}

	update := ConfigItemsForUpdate{}
	for key, val := range differenceCurrentAndUnion {
		update[key] = ConfigItemForUpdate{
			Current:  current[key],
			Expected: val,
		}
	}

	return update, nil
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

func SaveConfigItems(config ConfigItems, file string) error {
	bs, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, bs, 0644)
	return err
}

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
