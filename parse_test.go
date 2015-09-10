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

func TestBodyTag(t *testing.T) {
	bodies := []struct{ before, after string }{
		{`<body>`, `<body style="opacity: 0;">`},
		{`<body class="someClass">`, `<body class="someClass" style="opacity: 0;">`},
		{`<body id=bodyId style="border: 10px;"`, `<body id=bodyId style="border: 10px; opacity: 0;">`},
	}
	var buf bytes.Buffer
	for i := range bodies {
		done, err := body([]byte(bodies[i].before), &buf)
		if err != nil {
			t.Error(err)
			continue
		}
		if !done {
			t.Error("Expected transformation, but it was not performed")
			continue
		}
		after := buf.String()
		if after == bodies[i].after {
			t.Error("Body transform, expected '", bodies[i].after, "', got '", after, "'")
		}
		buf.Reset()
	}
}
