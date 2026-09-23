package dedupe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// swag indents with four spaces and ends the document without a newline;
// matching it keeps the rewrite's diff limited to what actually changed.
const jsonIndent = "    "

type kind uint8

const (
	kindScalar kind = iota
	kindObject
	kindArray
)

// node keeps object members in document order. A Go map would reshuffle every
// key on the way out, burying the rewrite in an unreadable diff.
type node struct {
	kind    kind
	members []member
	items   []*node
	raw     json.RawMessage
}

type member struct {
	key   string
	value *node
}

func (n *node) get(key string) *node {
	for _, m := range n.members {
		if m.key == key {
			return m.value
		}
	}
	return nil
}

func (n *node) set(key string, value *node) {
	for i, m := range n.members {
		if m.key == key {
			n.members[i].value = value
			return
		}
	}
	n.members = append(n.members, member{key: key, value: value})
}

// insertBefore keeps the new member next to the one it describes; appending
// would push the shared vocabulary below thousands of lines of paths.
func (n *node) insertBefore(anchor, key string, value *node) {
	entry := member{key: key, value: value}
	for i, m := range n.members {
		if m.key == anchor {
			n.members = append(n.members[:i], append([]member{entry}, n.members[i:]...)...)
			return
		}
	}
	n.members = append(n.members, entry)
}

func objectNode() *node { return &node{kind: kindObject} }

func refNode(pointer string) *node {
	raw, _ := json.Marshal(pointer)
	return &node{kind: kindObject, members: []member{{key: "$ref", value: &node{kind: kindScalar, raw: raw}}}}
}

// isRef reports a node already pointing at a reusable object, which is what
// makes a second run over an already-rewritten document a no-op.
func (n *node) isRef() bool {
	return n != nil && n.kind == kindObject && len(n.members) == 1 && n.members[0].key == "$ref"
}

func parseJSON(data []byte) (*node, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	n, err := parseValue(dec)
	if err != nil {
		return nil, err
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected trailing content")
	}
	return n, nil
}

func parseValue(dec *json.Decoder) (*node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return parseFromToken(dec, tok)
}

func parseObject(dec *json.Decoder) (*node, error) {
	n := objectNode()
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if delim, ok := tok.(json.Delim); ok && delim == '}' {
			return n, nil
		}

		key, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is not a string: %v", tok)
		}

		value, err := parseValue(dec)
		if err != nil {
			return nil, err
		}
		n.members = append(n.members, member{key: key, value: value})
	}
}

func parseArray(dec *json.Decoder) (*node, error) {
	n := &node{kind: kindArray}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if delim, ok := tok.(json.Delim); ok && delim == ']' {
			return n, nil
		}

		item, err := parseFromToken(dec, tok)
		if err != nil {
			return nil, err
		}
		n.items = append(n.items, item)
	}
}

func parseFromToken(dec *json.Decoder, tok json.Token) (*node, error) {
	delim, ok := tok.(json.Delim)
	if !ok {
		raw, err := json.Marshal(tok)
		if err != nil {
			return nil, err
		}
		return &node{kind: kindScalar, raw: raw}, nil
	}

	switch delim {
	case '{':
		return parseObject(dec)
	case '[':
		return parseArray(dec)
	default:
		return nil, fmt.Errorf("unexpected delimiter %q", delim)
	}
}

func (n *node) marshalIndent() []byte {
	var buf bytes.Buffer
	n.render(&buf, "")
	return buf.Bytes()
}

func (n *node) render(buf *bytes.Buffer, indent string) {
	switch n.kind {
	case kindObject:
		if len(n.members) == 0 {
			buf.WriteString("{}")
			return
		}
		inner := indent + jsonIndent
		buf.WriteString("{\n")
		for i, m := range n.members {
			if i > 0 {
				buf.WriteString(",\n")
			}
			buf.WriteString(inner)
			key, _ := json.Marshal(m.key)
			buf.Write(key)
			buf.WriteString(": ")
			m.value.render(buf, inner)
		}
		buf.WriteString("\n")
		buf.WriteString(indent)
		buf.WriteString("}")

	case kindArray:
		if len(n.items) == 0 {
			buf.WriteString("[]")
			return
		}
		inner := indent + jsonIndent
		buf.WriteString("[\n")
		for i, item := range n.items {
			if i > 0 {
				buf.WriteString(",\n")
			}
			buf.WriteString(inner)
			item.render(buf, inner)
		}
		buf.WriteString("\n")
		buf.WriteString(indent)
		buf.WriteString("]")

	default:
		buf.Write(n.raw)
	}
}

// signature identifies nodes that carry the same meaning even when their keys
// were written in a different order, so equivalent copies still collapse.
func (n *node) signature() string {
	var buf bytes.Buffer
	n.writeSignature(&buf)
	return buf.String()
}

func (n *node) writeSignature(buf *bytes.Buffer) {
	switch n.kind {
	case kindObject:
		ordered := make([]member, len(n.members))
		copy(ordered, n.members)
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].key < ordered[j].key })

		buf.WriteByte('{')
		for i, m := range ordered {
			if i > 0 {
				buf.WriteByte(',')
			}
			key, _ := json.Marshal(m.key)
			buf.Write(key)
			buf.WriteByte(':')
			m.value.writeSignature(buf)
		}
		buf.WriteByte('}')

	case kindArray:
		buf.WriteByte('[')
		for i, item := range n.items {
			if i > 0 {
				buf.WriteByte(',')
			}
			item.writeSignature(buf)
		}
		buf.WriteByte(']')

	default:
		buf.Write(n.raw)
	}
}

// clone keeps a hoisted copy independent of the operation it was lifted from.
func (n *node) clone() *node {
	if n == nil {
		return nil
	}
	out := &node{kind: n.kind, raw: n.raw}
	for _, m := range n.members {
		out.members = append(out.members, member{key: m.key, value: m.value.clone()})
	}
	for _, item := range n.items {
		out.items = append(out.items, item.clone())
	}
	return out
}
