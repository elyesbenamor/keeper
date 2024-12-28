package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Schema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	TTL         string                 `json:"ttl"`
	Fields      map[string]SchemaField `json:"fields"`
}

type SchemaField struct {
	Type     string                 `json:"type"`
	Required bool                   `json:"required"`
	Pattern  string                 `json:"pattern,omitempty"`
	MinLen   int                   `json:"minLength,omitempty"`
	MaxLen   int                   `json:"maxLength,omitempty"`
	Enum     []string              `json:"enum,omitempty"`
	Metadata map[string]string     `json:"metadata,omitempty"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage secret schemas",
}

var schemaAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		file, _ := cmd.Flags().GetString("file")

		// Read schema file
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		// Parse schema
		var schema Schema
		if err := json.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("invalid schema JSON: %w", err)
		}

		// Validate schema
		if schema.Name == "" {
			return fmt.Errorf("schema name is required")
		}
		if schema.Version == "" {
			return fmt.Errorf("schema version is required")
		}
		if len(schema.Fields) == 0 {
			return fmt.Errorf("schema must define at least one field")
		}

		// Save schema
		schemasDir := filepath.Join(configDir, "schemas")
		if err := os.MkdirAll(schemasDir, 0700); err != nil {
			return fmt.Errorf("failed to create schemas directory: %w", err)
		}

		schemaPath := filepath.Join(schemasDir, name+".json")
		if err := os.WriteFile(schemaPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write schema file: %w", err)
		}

		fmt.Printf("Schema '%s' added successfully\n", name)
		return nil
	},
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemasDir := filepath.Join(configDir, "schemas")
		entries, err := os.ReadDir(schemasDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No schemas found")
				return nil
			}
			return fmt.Errorf("failed to read schemas directory: %w", err)
		}

		fmt.Printf("%-20s %-10s %-30s\n", "NAME", "VERSION", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 60))

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}

			data, err := os.ReadFile(filepath.Join(schemasDir, entry.Name()))
			if err != nil {
				continue
			}

			var schema Schema
			if err := json.Unmarshal(data, &schema); err != nil {
				continue
			}

			name := strings.TrimSuffix(entry.Name(), ".json")
			fmt.Printf("%-20s %-10s %-30s\n", name, schema.Version, schema.Description)
		}

		return nil
	},
}

var schemaGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "View schema details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		schemaPath := filepath.Join(configDir, "schemas", name+".json")

		data, err := os.ReadFile(schemaPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("schema '%s' not found", name)
			}
			return fmt.Errorf("failed to read schema: %w", err)
		}

		var schema Schema
		if err := json.Unmarshal(data, &schema); err != nil {
			return fmt.Errorf("invalid schema JSON: %w", err)
		}

		// Pretty print schema
		formatted, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format schema: %w", err)
		}

		fmt.Println(string(formatted))
		return nil
	},
}

var schemaDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		schemaPath := filepath.Join(configDir, "schemas", name+".json")

		if err := os.Remove(schemaPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("schema '%s' not found", name)
			}
			return fmt.Errorf("failed to delete schema: %w", err)
		}

		fmt.Printf("Schema '%s' deleted successfully\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.AddCommand(schemaAddCmd)
	schemaCmd.AddCommand(schemaListCmd)
	schemaCmd.AddCommand(schemaGetCmd)
	schemaCmd.AddCommand(schemaDeleteCmd)

	schemaAddCmd.Flags().StringP("file", "f", "", "Schema file path")
	schemaAddCmd.MarkFlagRequired("file")
}
