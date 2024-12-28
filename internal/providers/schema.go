package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
)

// SchemaField represents a field in a schema
type SchemaField struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	MinLen      int                    `json:"minLength,omitempty"`
	MaxLen      int                    `json:"maxLength,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	Enum        []string              `json:"enum,omitempty"`
	Properties  map[string]SchemaField `json:"properties,omitempty"`
}

// Schema represents a JSON schema for validating secrets
type Schema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Fields      map[string]SchemaField `json:"fields"`
}

// LoadSchema loads a schema from a file
func LoadSchema(path string) (*Schema, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

// ValidateSecret validates a secret against its schema
func ValidateSecret(secret *Secret, schema *Schema) error {
	if schema == nil {
		return nil // No schema validation required
	}

	// Check required fields
	for name, field := range schema.Fields {
		if field.Required {
			if _, ok := secret.Metadata[name]; !ok {
				return fmt.Errorf("required field %s is missing", name)
			}
		}
	}

	// Validate field values
	for name, value := range secret.Metadata {
		field, ok := schema.Fields[name]
		if !ok {
			continue // Unknown field, skip validation
		}

		if err := validateField(name, value, field); err != nil {
			return err
		}
	}

	return nil
}

// validateField validates a single field value against its schema
func validateField(name, value string, field SchemaField) error {
	// Check string length
	if field.MinLen > 0 && len(value) < field.MinLen {
		return fmt.Errorf("field %s is too short (minimum length is %d)", name, field.MinLen)
	}
	if field.MaxLen > 0 && len(value) > field.MaxLen {
		return fmt.Errorf("field %s is too long (maximum length is %d)", name, field.MaxLen)
	}

	// Check pattern
	if field.Pattern != "" {
		pattern := regexp.MustCompile(field.Pattern)
		if !pattern.MatchString(value) {
			return fmt.Errorf("field %s does not match pattern %s", name, field.Pattern)
		}
	}

	// Check enum values
	if len(field.Enum) > 0 {
		valid := false
		for _, enum := range field.Enum {
			if value == enum {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("field %s must be one of: %v", name, field.Enum)
		}
	}

	return nil
}
