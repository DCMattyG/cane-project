package model

import (
	"bytes"
	"encoding/xml"
	"regexp"
	"strings"
)

// XMLNode Struct
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
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

// ScrubXML Function
func (x *XMLNode) ScrubXML() {
	if len(x.Nodes) > 0 {
		x.Content = []byte{}
	} else {
		contentString := string(x.Content)
		contentString = strings.Replace(contentString, "\n", "", -1)
		contentString = strings.Replace(contentString, "\r", "", -1)
		x.Content = []byte(contentString)
	}

	for i := 0; i < len(x.Nodes); i++ {
		x.Nodes[i].ScrubXML()
	}
}

// Marshal Function for XMLNode
func (x XMLNode) Marshal(args ...int) string {
	var xmlString string
	var depth int

	if len(args) > 0 {
		depth = args[0]
	} else {
		depth = 0
	}

	xmlString += strings.Repeat(" ", depth) + "<" + string(x.XMLName.Local)

	for _, attr := range x.Attrs {
		xmlString += " " + attr.Name.Local + "=\"" + attr.Value + "\""
	}

	xmlString += ">"

	if len(x.Content) > 0 && !IsCDATA(string(x.Content)) {
		xmlString += string(x.Content)
		xmlString += "</" + string(x.XMLName.Local) + ">\n"
	} else if IsCDATA(string(x.Content)) {
		var cdata XMLNode

		xmlString += "\n"
		xmlString += strings.Repeat(" ", (depth+2)) + "<![CDATA[\n"

		cdataString := string(x.Content)
		cdataString = strings.Replace(cdataString, "<![CDATA[", "", -1)
		cdataString = strings.Replace(cdataString, "]]>", "", -1)

		cdataErr := xml.Unmarshal([]byte(cdataString), &cdata)

		if cdataErr != nil {
			panic(cdataErr)
		}

		cdata.ScrubXML()

		xmlString += cdata.Marshal(depth + 4)
		xmlString += strings.Repeat(" ", (depth+2)) + "]]>\n"
		xmlString += strings.Repeat(" ", depth) + "</" + string(x.XMLName.Local) + ">\n"
	} else {
		xmlString += "\n"

		for _, node := range x.Nodes {
			xmlString += node.Marshal(depth + 2)
		}

		xmlString += strings.Repeat(" ", depth) + "</" + string(x.XMLName.Local) + ">\n"
	}

	return xmlString
}

// XMLVars Function for XMLNode
func (x *XMLNode) XMLVars() {
	for i := range x.Attrs {
		tempAttr := "{{var_" + string(x.Attrs[i].Name.Local) + "}}"
		x.Attrs[i].Value = tempAttr
	}

	contentString := string(x.Content)

	if contentString != "" {
		expression := regexp.MustCompile(`(<!--)[^-]+(-->)`)
		content := expression.ReplaceAllString(contentString, "")

		if IsCDATA(content) {
			newContent := strings.Replace(content, "<![CDATA[", "", 1)
			newContent = strings.Replace(newContent, "]]>", "", 1)

			var cdataNode XMLNode

			xmlErr := xml.Unmarshal([]byte(newContent), &cdataNode)

			if xmlErr != nil {
				panic(xmlErr)
			}

			cdataNode.XMLVars()

			cdataBytes, cdataErr := xml.Marshal(cdataNode)

			if cdataErr != nil {
				panic(cdataErr)
			}

			x.Content = []byte("<![CDATA[" + string(cdataBytes) + "]]>")
		} else {
			x.Content = []byte("{{var_" + string(x.XMLName.Local) + "}}")
		}
	}

	for i := 0; i < len(x.Nodes); i++ {
		x.Nodes[i].XMLVars()
	}
}

// UnmarshalXML Function
func (x *XMLNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	x.Attrs = start.Attr
	type node XMLNode

	return d.DecodeElement((*node)(x), &start)
}
