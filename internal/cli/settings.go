package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func settingsInstall(path, binary string) (string, error) {
	data, err := loadSettings(path)
	if err != nil {
		return "", err
	}

	keyword := hookKeyword(binary)

	for _, rule := range getSessionEndRules(data) {
		for _, h := range ruleHooks(rule) {
			if cmd, ok := hookCommand(h); ok && strings.Contains(cmd, keyword) {
				return "already_configured", nil
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
		return "", err
	}

	return "installed", nil
}

func settingsUninstall(path, binary string) (string, error) {
	data, err := loadSettings(path)
	if err != nil {
		return "", err
	}

	keyword := hookKeyword(binary)

	rules := getSessionEndRules(data)
	if len(rules) == 0 {
		return "not_found", nil
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
		return "", err
	}

	return "uninstalled", nil
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
