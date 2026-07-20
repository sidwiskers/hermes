// Package botapi owns the checked-in representation of the official Telegram
// Bot API schema and the tools that derive it from Telegram's documentation.
package botapi

import (
	"encoding/json"
	"fmt"
	"os"
)

const OfficialSource = "https://core.telegram.org/bots/api"

type Manifest struct {
	Version      string   `json:"version"`
	Released     string   `json:"released"`
	Source       string   `json:"source"`
	SourceSHA256 string   `json:"source_sha256"`
	Methods      []Method `json:"methods"`
	Objects      []Object `json:"objects"`
	Unions       []Union  `json:"unions"`
}

type Method struct {
	Name       string  `json:"name"`
	Anchor     string  `json:"anchor"`
	Parameters []Field `json:"parameters"`
}

type Object struct {
	Name   string  `json:"name"`
	Anchor string  `json:"anchor"`
	Fields []Field `json:"fields"`
}

type Union struct {
	Name         string   `json:"name"`
	Anchor       string   `json:"anchor"`
	Variants     []string `json:"variants"`
	Alternatives []string `json:"alternatives,omitempty"`
}

type Field struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Stats struct {
	Methods      int `json:"methods"`
	Parameters   int `json:"parameters"`
	Objects      int `json:"objects"`
	ObjectFields int `json:"object_fields"`
	Unions       int `json:"unions"`
	Variants     int `json:"variants"`
}

func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var schema Manifest
	if err := json.Unmarshal(data, &schema); err != nil {
		return Manifest{}, err
	}
	if err := schema.Validate(); err != nil {
		return Manifest{}, err
	}
	return schema, nil
}

func Marshal(schema Manifest) ([]byte, error) {
	if err := schema.Validate(); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (schema Manifest) Stats() Stats {
	result := Stats{Methods: len(schema.Methods), Objects: len(schema.Objects), Unions: len(schema.Unions)}
	for _, method := range schema.Methods {
		result.Parameters += len(method.Parameters)
	}
	for _, object := range schema.Objects {
		result.ObjectFields += len(object.Fields)
	}
	for _, union := range schema.Unions {
		result.Variants += len(union.Variants)
	}
	return result
}

func (schema Manifest) Validate() error {
	if schema.Version == "" || schema.Source == "" || schema.SourceSHA256 == "" {
		return fmt.Errorf("botapi: incomplete manifest metadata")
	}
	if err := validateNamed("method", len(schema.Methods), func(index int) (string, string) {
		return schema.Methods[index].Name, schema.Methods[index].Anchor
	}); err != nil {
		return err
	}
	if err := validateNamed("object", len(schema.Objects), func(index int) (string, string) {
		return schema.Objects[index].Name, schema.Objects[index].Anchor
	}); err != nil {
		return err
	}
	if err := validateNamed("union", len(schema.Unions), func(index int) (string, string) {
		return schema.Unions[index].Name, schema.Unions[index].Anchor
	}); err != nil {
		return err
	}
	for _, method := range schema.Methods {
		if err := validateFields("method "+method.Name, method.Parameters); err != nil {
			return err
		}
	}
	for _, object := range schema.Objects {
		if err := validateFields("object "+object.Name, object.Fields); err != nil {
			return err
		}
	}
	for _, union := range schema.Unions {
		if len(union.Variants) == 0 {
			return fmt.Errorf("botapi: union %s has no variants", union.Name)
		}
	}
	return nil
}

func validateNamed(kind string, count int, at func(int) (string, string)) error {
	seenNames := make(map[string]struct{}, count)
	seenAnchors := make(map[string]struct{}, count)
	for index := 0; index < count; index++ {
		name, anchor := at(index)
		if name == "" || anchor == "" {
			return fmt.Errorf("botapi: %s has an empty name or anchor", kind)
		}
		if _, exists := seenNames[name]; exists {
			return fmt.Errorf("botapi: duplicate %s name %s", kind, name)
		}
		if _, exists := seenAnchors[anchor]; exists {
			return fmt.Errorf("botapi: duplicate %s anchor %s", kind, anchor)
		}
		seenNames[name] = struct{}{}
		seenAnchors[anchor] = struct{}{}
	}
	return nil
}

func validateFields(owner string, fields []Field) error {
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if field.Name == "" || field.Type == "" {
			return fmt.Errorf("botapi: %s has an incomplete field", owner)
		}
		if _, exists := seen[field.Name]; exists {
			return fmt.Errorf("botapi: %s has duplicate field %s", owner, field.Name)
		}
		seen[field.Name] = struct{}{}
	}
	return nil
}
