package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var (
	tagB     = []byte(`<!--  CONTENT ELEMENT, `)
	uidB     = []byte(`uid:`)
	endB     = []byte(`-->`)
	hideB    = []byte(`<div style="opacity: 0;">`)
	hideEndB = []byte(`</div>`)
)

type element interface {
	seed(uid int)
	writeTo(w io.Writer) error
}

// Pure HTML element, no seeding
type text []byte

func (t *text) String() string {
	return "[simple text element]"
}

func (t *text) seed(uid int) {
}

func (t *text) writeTo(w io.Writer) error {
	_, err := w.Write([]byte(*t))
	return err
}

// Can be shown or hidden. If hidden, only non-text children are shown.
type marker struct {
	uid    int
	name   string
	active bool
	text   []byte
}

func (m *marker) String() string {
	return fmt.Sprintf("[mark %d/%s]", m.uid, m.name)
}

func (m *marker) seed(uid int) {
	if uid == m.uid {
		m.active = true
	}
}

func (m *marker) writeTo(w io.Writer) error {
	if !m.active {
		if _, err := w.Write(hideB); err != nil {
			return err
		}
	}
	if _, err := w.Write(m.text); err != nil {
		return err
	}
	if !m.active {
		if _, err := w.Write(hideEndB); err != nil {
			return err
		}
	}
	return nil
}

type parser struct {
	elements []element
}

func (p *parser) seed(uid int) {
	for i := range p.elements {
		p.elements[i].seed(uid)
	}
}

func (p *parser) parse(r io.Reader) error {
	var pos int
	wb := &bytes.Buffer{}
	if _, err := io.Copy(wb, r); err != nil {
		return err
	}
	buf := wb.Bytes()
	elements := make([]element, 0)
	for {
		i := bytes.Index(buf[pos:], tagB)
		if i == -1 {
			// after last tag
			t := text(buf[pos:])
			elements = append(elements, &t)
			break
		}
		// before new tag
		i = i + pos
		// only if not empty
		if pos < i {
			t := text(buf[pos:i])
			elements = append(elements, &t)
		}
		// tag and its contents
		m := &marker{}
		n, err := p.parseTag(m, buf[i:])
		if err != nil {
			return err
		}
		pos = i + n
		i = bytes.Index(buf[pos:], tagB)
		if i == -1 {
			return fmt.Errorf("%s: Expected closing tag not found", m)
		}
		i = i + pos
		// set content of this marker
		m.text = buf[pos:i]
		n, err = p.parseTag(m, buf[i:])
		if err != nil {
			return err
		}
		// append marker content element
		elements = append(elements, m)
		pos = i + n
	}
	p.elements = elements
	return nil
}

func (p *parser) parseTag(m *marker, buf []byte) (int, error) {
	i, j := len(tagB), 0
	// skip spaces
	for ; buf[i] == ' '; i++ {
	}
	if !bytes.HasPrefix(buf[i:], uidB) {
		return 0, errors.New("Invalid tag, no 'uid:' found")
	}
	i = i + len(uidB)
	// end of number
	for j = i; buf[j] != '/'; j++ {
	}
	n, err := strconv.ParseInt(string(buf[i:j]), 10, 32)
	if err != nil {
		return 0, err
	}
	m.uid = int(n)
	// end of type
	for i = j; buf[i] != ' '; i++ {
	}
	m.name = string(buf[j+1 : i])
	// end of tag
	// XXX: [begin] or [end] not checked.
	j = bytes.Index(buf[i:], endB)
	if j == -1 {
		return 0, errors.New("Invalid unended tag")
	}
	return j + i + len(endB), nil
}

func main() {
	var p parser
	if err := p.parse(os.Stdin); err != nil {
		log.Fatal(err)
	}
	p.seed(32987)
	for _, el := range p.elements {
		el.writeTo(os.Stdout)
	}
}
