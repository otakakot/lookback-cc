package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: settings <install|uninstall|check> <binary-path>")
		os.Exit(1)
	}

	action := os.Args[1]
	binary := os.Args[2]

	settingsPath := os.Getenv("SETTINGS_PATH")
	if settingsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
			os.Exit(1)
		}
		settingsPath = home + "/.claude/settings.json"
	}

	switch action {
	case "check":
		if hasHook(settingsPath, binary) {
			fmt.Println("found")
			os.Exit(0)
		}
		fmt.Println("not_found")
		os.Exit(1)
	case "install":
		os.Exit(install(settingsPath, binary))
	case "uninstall":
		os.Exit(uninstall(settingsPath, binary))
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown action %q (use install, uninstall, or check)\n", action)
		os.Exit(1)
	}
}

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

func loadSettings(path string) (*orderedMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &orderedMap{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var data orderedMap
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	return &data, nil
}

func saveSettings(path string, data *orderedMap) error {
	compact, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, compact, "", "  "); err != nil {
		return fmt.Errorf("indent: %w", err)
	}
	buf.WriteByte('\n')

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func hasHook(path, binary string) bool {
	data, err := loadSettings(path)
	if err != nil {
		return false
	}
	keyword := hookKeyword(binary)
	for _, rule := range getSessionEndRules(data) {
		for _, h := range ruleHooks(rule) {
			if cmd, ok := hookCommand(h); ok && strings.Contains(cmd, keyword) {
				return true
			}
		}
	}
	return false
}

func install(path, binary string) int {
	data, err := loadSettings(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	keyword := hookKeyword(binary)
	for _, rule := range getSessionEndRules(data) {
		for _, h := range ruleHooks(rule) {
			if cmd, ok := hookCommand(h); ok && strings.Contains(cmd, keyword) {
				fmt.Println("already_configured")
				return 0
			}
		}
	}

	newRule := &orderedMap{entries: []entry{
		{"hooks", []any{
			&orderedMap{entries: []entry{
				{"type", "command"},
				{"command", binary},
			}},
		}},
	}}

	hooks := ensureHooks(data)
	rules := getSessionEndRules(data)
	rules = append(rules, newRule)
	hooks.set("SessionEnd", rules)

	if err := saveSettings(path, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("installed")
	return 0
}

func uninstall(path, binary string) int {
	data, err := loadSettings(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	keyword := hookKeyword(binary)
	rules := getSessionEndRules(data)
	if len(rules) == 0 {
		fmt.Println("not_found")
		return 0
	}

	var filtered []any
	for _, rule := range rules {
		match := false
		for _, h := range ruleHooks(rule) {
			if cmd, ok := hookCommand(h); ok && strings.Contains(cmd, keyword) {
				match = true
				break
			}
		}
		if !match {
			filtered = append(filtered, rule)
		}
	}

	hooks := ensureHooks(data)
	if len(filtered) > 0 {
		hooks.set("SessionEnd", filtered)
	} else {
		hooks.delete("SessionEnd")
	}
	if hooks.len() == 0 {
		data.delete("hooks")
	}

	if err := saveSettings(path, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("uninstalled")
	return 0
}

func hookKeyword(binary string) string {
	return filepath.Base(binary)
}

func ensureHooks(data *orderedMap) *orderedMap {
	v, ok := data.get("hooks")
	if ok {
		if hooks, ok := v.(*orderedMap); ok {
			return hooks
		}
	}
	hooks := &orderedMap{}
	data.set("hooks", hooks)
	return hooks
}

func getSessionEndRules(data *orderedMap) []any {
	v, ok := data.get("hooks")
	if !ok {
		return nil
	}
	hooks, ok := v.(*orderedMap)
	if !ok {
		return nil
	}
	v, ok = hooks.get("SessionEnd")
	if !ok {
		return nil
	}
	rules, ok := v.([]any)
	if !ok {
		return nil
	}
	return rules
}

func ruleHooks(rule any) []any {
	m, ok := rule.(*orderedMap)
	if !ok {
		return nil
	}
	v, ok := m.get("hooks")
	if !ok {
		return nil
	}
	hs, ok := v.([]any)
	if !ok {
		return nil
	}
	return hs
}

func hookCommand(h any) (string, bool) {
	m, ok := h.(*orderedMap)
	if !ok {
		return "", false
	}
	v, ok := m.get("command")
	if !ok {
		return "", false
	}
	cmd, ok := v.(string)
	return cmd, ok
}

var _ json.Marshaler = (*orderedMap)(nil)
var _ json.Unmarshaler = (*orderedMap)(nil)
