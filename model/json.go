package model

import (
	"encoding/json"
	"reflect"
)

// JSONNode Type
type JSONNode map[string]interface{}

// IsJSON Function
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// JSONVars Function for JSONNode
func (j JSONNode) JSONVars() {
	for key, val := range j {
		switch valType := reflect.ValueOf(val).Kind(); valType {
		case reflect.Map:
			tempKey := JSONNode(j[key].(map[string]interface{}))
			tempKey.JSONVars()
			j[key] = tempKey
		default:
			j[key] = "{{var_" + key + "}}"
		}
	}
}
