package main

import (
	"bytes"
	"testing"
)

const html = `
Pre tags
<!--  CONTENT ELEMENT, uid:1/first [begin] -->
Tag 1 content
<!--  CONTENT ELEMENT, uid:1/first [end] -->
<!--  CONTENT ELEMENT, uid:2/following [begin] -->
Tag 2 content
<!--  CONTENT ELEMENT, uid:2/following [end] -->
<!--  CONTENT ELEMENT, uid:3/no-content [begin] --><!--  CONTENT ELEMENT, uid:3/no-content [end] -->
Trailing contents
`

func TestParsing(t *testing.T) {
	var p parser
	buf := bytes.NewBuffer([]byte(html))
	if err := p.parse(buf); err != nil {
		t.Fatal(err)
	}
	if len(p.elements) != 7 {
		t.Error("Expected 7 elements, got ", len(p.elements))
	}
	m, ok := p.elements[5].(*marker)
	if !ok {
		t.Fatal("Expected marker for empty element")
	}
	if string(m.text) != "" {
		t.Error("Empty element is not empty: '", string(m.text), "'")
	}
	// TODO: test content of other elements.
}
