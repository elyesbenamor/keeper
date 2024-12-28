package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// FieldType represents the type of a secret field
type FieldType string

const (
	TypeString   FieldType = "string"
	TypeNumber   FieldType = "number"
	TypeBoolean  FieldType = "boolean"
	TypePassword FieldType = "password"
)

// Field represents a field in a secret schema
type Field struct {
	Type        FieldType          `json:"type"`
	Required    bool              `json:"required"`
	Pattern     string            `json:"pattern,omitempty"`
	MinLength   *int              `json:"minLength,omitempty"`
	MaxLength   *int              `json:"maxLength,omitempty"`
	Enum        []string          `json:"enum,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Schema represents a secret schema
type Schema struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Fields      map[string]Field `json:"fields"`
	TTL         *time.Duration   `json:"ttl,omitempty"`
}

// Validate validates a secret value against the schema
func (s *Schema) Validate(value map[string]interface{}) error {
	// Check required fields
	for name, field := range s.Fields {
		if field.Required {
			if _, ok := value[name]; !ok {
				return fmt.Errorf("missing required field: %s", name)
			}
		}
	}

	// Validate each field
	for name, val := range value {
		field, ok := s.Fields[name]
		if !ok {
			return fmt.Errorf("unknown field: %s", name)
		}

		if err := field.validate(val); err != nil {
			return fmt.Errorf("invalid field %s: %w", name, err)
		}
	}

	return nil
}

// validate validates a single field value
func (f *Field) validate(value interface{}) error {
	// Check type
	switch f.Type {
	case TypeString, TypePassword:
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		// Check length constraints
		if f.MinLength != nil && len(str) < *f.MinLength {
			return fmt.Errorf("string length %d is less than minimum %d", len(str), *f.MinLength)
		}
		if f.MaxLength != nil && len(str) > *f.MaxLength {
			return fmt.Errorf("string length %d is greater than maximum %d", len(str), *f.MaxLength)
		}

		// Check pattern
		if f.Pattern != "" {
			matched, err := regexp.MatchString(f.Pattern, str)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
			if !matched {
				return fmt.Errorf("value does not match pattern: %s", f.Pattern)
			}
		}

		// Check enum
		if len(f.Enum) > 0 {
			valid := false
			for _, enum := range f.Enum {
				if str == enum {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("value must be one of: %v", f.Enum)
			}
		}

	case TypeNumber:
		switch value.(type) {
		case float64, int, int64:
			// Valid number type
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

	case TypeBoolean:
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	}

	return nil
}

// LoadSchema loads a schema from JSON data
func LoadSchema(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}
	return &schema, nil
}

// GetExpirationTime returns when the secret should expire based on the schema TTL
func (s *Schema) GetExpirationTime() *time.Time {
	if s.TTL == nil {
		return nil
	}
	t := time.Now().Add(*s.TTL)
	return &t
}
