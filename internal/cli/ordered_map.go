package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// orderedMap preserves JSON key order.
type orderedMap struct {
	entries []entry
}

type entry struct {
	key   string
	value any
}

func (m *orderedMap) get(key string) (any, bool) {
	for _, e := range m.entries {
		if e.key == key {
			return e.value, true
		}
	}

	return nil, false
}

func (m *orderedMap) set(key string, value any) {
	for i, e := range m.entries {
		if e.key == key {
			m.entries[i].value = value
			return
		}
	}

	m.entries = append(m.entries, entry{key, value})
}

func (m *orderedMap) delete(key string) {
	for i, e := range m.entries {
		if e.key == key {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			return
		}
	}
}

func (m *orderedMap) len() int {
	return len(m.entries)
}

func (m *orderedMap) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))

	tok, err := dec.Token()
	if err != nil {
		return err
	}

	if tok != json.Delim('{') {
		return fmt.Errorf("expected '{', got %v", tok)
	}

	m.entries = nil

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}

		key := keyTok.(string)

		val, err := decodeValue(dec)
		if err != nil {
			return err
		}

		m.entries = append(m.entries, entry{key, val})
	}

	// consume closing '}'
	_, err = dec.Token()

	return err
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := &orderedMap{}

			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}

				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}

				obj.entries = append(obj.entries, entry{keyTok.(string), val})
			}
			// consume closing '}'
			if _, err := dec.Token(); err != nil {
				return nil, err
			}

			return obj, nil
		case '[':
			var arr []any

			for dec.More() {
				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}

				arr = append(arr, val)
			}
			// consume closing ']'
			if _, err := dec.Token(); err != nil {
				return nil, err
			}

			return arr, nil
		}
	default:
		return t, nil
	}

	return nil, fmt.Errorf("unexpected token: %v", tok)
}

func (m *orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, e := range m.entries {
		if i > 0 {
			buf.WriteByte(',')
		}

		key, err := json.Marshal(e.key)
		if err != nil {
			return nil, err
		}

		buf.Write(key)
		buf.WriteByte(':')

		val, err := json.Marshal(e.value)
		if err != nil {
			return nil, err
		}

		buf.Write(val)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

var _ json.Marshaler = (*orderedMap)(nil)
var _ json.Unmarshaler = (*orderedMap)(nil)
