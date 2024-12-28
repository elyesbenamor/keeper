# Keeper Advanced Features

This document describes the advanced features of Keeper, including schema validation, backup/restore, and batch operations.

## Table of Contents
- [Schema Validation](#schema-validation)
- [Backup and Restore](#backup-and-restore)
- [Batch Operations](#batch-operations)
- [Secret Expiration](#secret-expiration)

## Schema Validation

Schema validation allows you to define and enforce structure and validation rules for your secrets. This ensures consistency and helps prevent errors.

### Schema Structure

Schemas are defined in JSON format with the following structure:

```json
{
  "name": "api_key",
  "description": "API Key Schema",
  "version": "1.0",
  "ttl": "720h",
  "fields": {
    "value": {
      "type": "string",
      "required": true,
      "pattern": "^[A-Za-z0-9]{32}$",
      "metadata": {
        "description": "API Key value"
      }
    }
  }
}
```

### Field Types

- `string`: Text values
- `number`: Numeric values (integers or floats)
- `boolean`: True/false values
- `password`: Sensitive string values (treated with extra care)

### Validation Rules

- `required`: Field must be present
- `pattern`: Regular expression pattern for string validation
- `minLength`/`maxLength`: String length constraints
- `enum`: List of allowed values
- `metadata`: Additional field metadata

### Schema Management Commands

```bash
# Add a new schema
kpr schema add api_key --file api_key_schema.json

# List all schemas
kpr schema list

# View schema details
kpr schema get api_key

# Delete a schema
kpr schema delete api_key
```

### Using Schemas with Secrets

```bash
# Set a secret with schema validation
kpr set my_api_key "abc123..." --schema api_key

# Set a secret from environment variable with schema
export DB_PASSWORD="secret123"
kpr set db_password --from-env --schema db_credentials
```

## Backup and Restore

Keeper provides robust backup and restore functionality to help you manage and migrate your secrets.

### Backup Features

- Multiple formats (JSON, YAML)
- Optional compression
- Optional encryption
- Version history support

### Backup Commands

```bash
# Create an encrypted JSON backup with version history
kpr backup secrets.json --versions --compress --encrypt

# Create an unencrypted YAML backup
kpr backup secrets.yaml --format yaml --encrypt=false

# Create a compressed backup
kpr backup secrets.json --compress
```

### Restore Commands

```bash
# Restore secrets, skipping existing ones
kpr restore secrets.json

# Restore secrets and overwrite existing ones
kpr restore secrets.json --overwrite

# Restore secrets with version history
kpr restore secrets.json --versions
```

## Batch Operations

Batch operations allow you to efficiently manage multiple secrets at once.

### Batch Commands

```bash
# Get multiple secrets
kpr batch get "secret1,secret2,secret3"

# Get multiple secrets and save to file
kpr batch get "secret1,secret2,secret3" --file secrets.json

# Set multiple secrets from file
kpr batch set "secret1,secret2,secret3" --file secrets.json
```

### Batch File Format

```json
{
  "secret1": {
    "value": "value1",
    "metadata": {
      "env": "prod",
      "app": "myapp"
    }
  },
  "secret2": {
    "value": "value2",
    "metadata": {
      "env": "staging"
    }
  }
}
```

## Secret Expiration

Secrets can be configured to automatically expire after a certain time.

### Expiration Methods

1. **Schema-based TTL**:
   ```json
   {
     "name": "temporary_token",
     "ttl": "24h",
     "fields": {
       "value": {
         "type": "string",
         "required": true
       }
     }
   }
   ```

2. **Manual Expiration**:
   ```bash
   # Set a secret that expires in 24 hours
   kpr set temp_secret "value" --expire 24h

   # Set a secret that expires in 30 days
   kpr set temp_secret "value" --expire 720h
   ```

### Cleanup

```bash
# Remove expired secrets
kpr cleanup
```

## Example Schemas

### API Key Schema
```json
{
  "name": "api_key",
  "description": "API Key Schema",
  "version": "1.0",
  "ttl": "8760h",
  "fields": {
    "value": {
      "type": "string",
      "required": true,
      "pattern": "^[A-Za-z0-9]{32,64}$",
      "metadata": {
        "description": "API Key value"
      }
    }
  }
}
```

### Database Credentials Schema
```json
{
  "name": "db_credentials",
  "description": "Database Credentials Schema",
  "version": "1.0",
  "ttl": "720h",
  "fields": {
    "username": {
      "type": "string",
      "required": true,
      "minLength": 3,
      "maxLength": 64
    },
    "password": {
      "type": "password",
      "required": true,
      "minLength": 8,
      "pattern": "^(?=.*[A-Za-z])(?=.*\\d)[A-Za-z\\d@$!%*#?&]{8,}$",
      "metadata": {
        "description": "Must contain at least one letter and one number"
      }
    },
    "host": {
      "type": "string",
      "required": true,
      "pattern": "^[a-zA-Z0-9.-]+$"
    },
    "port": {
      "type": "number",
      "required": true
    },
    "ssl": {
      "type": "boolean",
      "required": true
    }
  }
}
```

### OAuth Token Schema
```json
{
  "name": "oauth_token",
  "description": "OAuth Token Schema",
  "version": "1.0",
  "ttl": "1h",
  "fields": {
    "access_token": {
      "type": "string",
      "required": true
    },
    "refresh_token": {
      "type": "string",
      "required": true
    },
    "token_type": {
      "type": "string",
      "required": true,
      "enum": ["Bearer", "MAC"]
    },
    "expires_in": {
      "type": "number",
      "required": true
    }
  }
}
```

## Best Practices

1. **Schema Design**:
   - Use descriptive names and versions
   - Include proper validation rules
   - Set appropriate TTLs
   - Add helpful metadata

2. **Backup Strategy**:
   - Regular backups
   - Encrypt sensitive backups
   - Store backups securely
   - Test restore process

3. **Secret Management**:
   - Use schemas for critical secrets
   - Set appropriate expiration times
   - Regular cleanup of expired secrets
   - Use batch operations for bulk updates

4. **Security**:
   - Use strong patterns for passwords
   - Enable encryption for sensitive backups
   - Regular rotation of long-lived secrets
   - Proper access control to backup files
