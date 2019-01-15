package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

// XMLNode Struct
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
}

// JSONNode Struct
type JSONNode struct {
	Node map[string]interface{}
}

// RespondwithJSON write json response format
func RespondwithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	fmt.Println(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// RespondWithError return error message
func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondwithJSON(w, code, map[string]string{"message": msg})
}

// UnmarshalJSON Function
func UnmarshalJSON(data []byte, target interface{}) error {
	err := json.Unmarshal(data, &target)

	if err != nil {
		return err
	}

	return nil
}

// StringInSlice Function
func StringInSlice(a []string, b string) bool {
	for _, i := range a {
		if b == i {
			return true
		}
	}
	return false
}

// IsXML Function
func IsXML(s string) bool {
	buf := bytes.NewBuffer([]byte(s))
	dec := xml.NewDecoder(buf)

	var n XMLNode

	err := dec.Decode(&n)

	if err != nil {
		return false
	}

	return true
}

// IsJSON Function
func IsJSON(s string) bool {
	var js map[string]interface{}

	err := json.Unmarshal([]byte(s), &js)

	if err != nil {
		return false
	}

	return true
}
