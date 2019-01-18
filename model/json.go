package model

import (
	"encoding/json"
	"reflect"
	"strings"
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

// Marshal Function for JSONNode
func (j JSONNode) Marshal(args ...int) string {
	prefix := ""
	indent := "    "

	if len(args) == 1 {
		prefix = strings.Repeat(" ", args[0])
	} else if len(args) == 2 {
		prefix = strings.Repeat(" ", args[0])
		indent = strings.Repeat(" ", args[1])
	}

	jsonBytes, jsonErr := json.MarshalIndent(j, prefix, indent)
	jsonString := string(jsonBytes)

	if jsonErr != nil {
		panic(jsonErr)
	}

	return jsonString
}
