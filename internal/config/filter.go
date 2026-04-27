package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Filter interface{ kind() string }

type ScaleFilter struct {
	Width  string `yaml:"width"`
	Height string `yaml:"height"`
}

func (s *ScaleFilter) kind() string { return "scale" }

type TimestampFilter struct{}

func (t *TimestampFilter) kind() string { return "timestamp" }

type CustomFilter struct {
	String string `yaml:"string"`
}

func (c *CustomFilter) kind() string { return "custom" }

type PhraseFilter struct{}

func (p *PhraseFilter) kind() string { return "phrase" }

// FilterEntry is the YAML-dispatchable wrapper around a Filter.
// Scenario.Filters holds []FilterEntry so the custom unmarshaler runs per element.
type FilterEntry struct {
	Filter
}

func (fe *FilterEntry) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]string
	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("filter: invalid structure: %w", err)
	}

	typVal, ok := raw["type"]
	if !ok {
		return fmt.Errorf("filter: missing 'type' field")
	}

	switch typVal {
	case "scale":
		var f ScaleFilter
		if err := value.Decode(&f); err != nil {
			return fmt.Errorf("filter scale: %w", err)
		}
		if f.Width == "" || f.Height == "" {
			return fmt.Errorf("filter scale: 'width' and 'height' are required")
		}
		fe.Filter = &f

	case "timestamp":
		allowed := map[string]bool{"type": true}
		for k := range raw {
			if !allowed[k] {
				return fmt.Errorf("filter timestamp: unexpected field %q", k)
			}
		}
		fe.Filter = &TimestampFilter{}

	case "custom":
		var f CustomFilter
		if err := value.Decode(&f); err != nil {
			return fmt.Errorf("filter custom: %w", err)
		}
		if f.String == "" {
			return fmt.Errorf("filter custom: 'string' is required")
		}
		fe.Filter = &f

	case "phrase":
		allowed := map[string]bool{"type": true}
		for k := range raw {
			if !allowed[k] {
				return fmt.Errorf("filter phrase: unexpected field %q", k)
			}
		}
		fe.Filter = &PhraseFilter{}

	default:
		return fmt.Errorf("filter: unknown type %q", typVal)
	}

	return nil
}
