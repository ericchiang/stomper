package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// merge joins the 'config' field of two json metadata files.
func merge(oldJSON, newJSON []byte) ([]byte, error) {
	var oldData, newData map[string]interface{}
	if err := json.Unmarshal(oldJSON, &oldData); err != nil {
		return nil, fmt.Errorf("failed to parse old JSON meta: %v", err)
	}
	if err := json.Unmarshal(newJSON, &newData); err != nil {
		return nil, fmt.Errorf("failed to parse new JSON meta: %v", err)
	}

	config := func(data map[string]interface{}) (map[string]interface{}, bool) {
		if config, ok := data["config"].(map[string]interface{}); ok {
			return config, true
		}
		return nil, false
	}

	oldConfig, ok := config(oldData)
	if !ok {
		return nil, errors.New("no 'config' field")
	}
	newConfig, ok := config(newData)
	if !ok {
		return nil, errors.New("no 'config' field")
	}

	if err := mergeJSON(oldConfig, newConfig); err != nil {
		return nil, fmt.Errorf("merge failed: %v", err)
	}

	newData["config"] = oldConfig
	data, err := json.Marshal(&newData)
	if err != nil {
		return nil, fmt.Errorf("merge failed: %v", err)
	}
	return data, nil
}

func mergeJSON(oldData, newData map[string]interface{}) error {
	for field, oldVal := range oldData {
		if newVal, ok := newData[field]; ok {
			if oldVal == nil {
				oldData[field] = newVal
			} else {
				switch newVal := newVal.(type) {
				default:
					if newVal != nil {
						oldData[field] = newVal
					}
				case map[string]interface{}:
					if oldVal, ok := oldVal.(map[string]interface{}); ok {
						if err := mergeJSON(oldVal, newVal); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("field %s has conflicting types between version", field)
					}
				}
			}
		}
	}
	for field, val := range newData {
		if _, ok := oldData[field]; !ok {
			oldData[field] = val
		}
	}
	return nil
}
