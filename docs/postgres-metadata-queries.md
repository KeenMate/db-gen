# PostgreSQL Metadata Queries Used by db-gen

This document contains the SQL queries that db-gen uses to extract function/procedure metadata from PostgreSQL.

## Overview

db-gen queries PostgreSQL's `information_schema` to get information about:
1. Functions and procedures (routines)
2. Input parameters
3. Return types (output parameters)

## 1. Get Functions in Schema

This query retrieves all functions and procedures in a specific schema.

**Location in code:** `private/dbGen/routinesProvider.go` lines 124-139

**Parameters:**
- `$1` - Schema name (e.g., `'public'`, `'internal'`, `'auth'`)

```sql
SELECT
    row_number() OVER (PARTITION BY routine_schema, routine_name),
    r.routine_schema::text,
    r.routine_name::text,
    r.specific_name::text,
    COALESCE(r.data_type, '') AS data_type,
    COALESCE(r.type_udt_schema::text, '') AS type_udt_schema,
    COALESCE(r.type_udt_name::text, '') AS type_udt_name,
    COALESCE(param_count, 0) AS param_count,
    CASE
        WHEN r.data_type IS NULL THEN 'procedure'
        ELSE 'function'
    END AS func_type
FROM information_schema.routines r
LEFT JOIN (
    SELECT
        specific_schema,
        specific_name,
        COUNT(*) AS param_count
    FROM information_schema.parameters p
    GROUP BY specific_schema, specific_name
) p ON p.specific_schema = r.specific_schema
   AND p.specific_name = r.specific_name
WHERE r.specific_schema = $1
ORDER BY routine_schema, routine_name;
```

**Returns:**
- `routine_schema` - Schema name
- `routine_name` - Function name (without parameters)
- `specific_name` - PostgreSQL-specific internal name (unique identifier)
- `data_type` - Return data type (NULL for procedures)
- `type_udt_schema` - Schema of user-defined return type
- `type_udt_name` - Name of user-defined return type (for composite/TABLE returns)
- `param_count` - Number of parameters
- `func_type` - Either 'function' or 'procedure'

## 2. Get Parameters and Return Types

This query retrieves both input parameters and return type columns for a specific function.

**Location in code:** `private/dbGen/routinesProvider.go` lines 151-172

**Parameters:**
- `$1` - Routine schema
- `$2` - Specific name (internal PostgreSQL name)
- `$3` - UDT type name (for return type lookup)
- `$4` - UDT type schema (for return type lookup)

```sql
-- Part 1: Input Parameters
SELECT
    ordinal_position::int,
    parameter_name::text,
    parameter_mode::text,
    udt_name::text,
    false AS is_nullable,  -- Always false for input params!
    parameter_default IS NOT NULL AS is_optional
FROM information_schema.parameters
WHERE specific_schema = $1
  AND specific_name = $2

UNION

-- Part 2: Return Columns (TABLE return type)
SELECT
    c.ordinal_position::int,
    c.column_name::text,
    'OUT',
    c.udt_name::text,
    c.is_nullable = 'YES',  -- Actual nullable info from metadata
    true
FROM information_schema.columns c
WHERE c.table_name = $3
  AND c.table_schema = COALESCE($4, 'public')

UNION

-- Part 3: Return Attributes (Composite type return)
SELECT
    a.ordinal_position::int,
    a.attribute_name::text,
    'OUT',
    a.attribute_udt_name::text,
    is_nullable = 'YES',  -- Actual nullable info from metadata
    true
FROM information_schema.attributes a
WHERE a.udt_name = $3
  AND a.udt_schema = COALESCE($4, 'public')

ORDER BY ordinal_position;
```

**Returns:**
- `ordinal_position` - Position in parameter/return list (1-based)
- `parameter_name` / `column_name` / `attribute_name` - Name
- `parameter_mode` - 'IN' or 'OUT'
- `udt_name` / `attribute_udt_name` - PostgreSQL type (int4, text, jsonb, etc.)
- `is_nullable` - Whether the parameter/column can be NULL
- `is_optional` - Whether parameter has a DEFAULT value (only for IN params)

## Important Notes

### Input Parameter Nullability

**CRITICAL:** Line 156 hardcodes `false as is_nullable` for input parameters!

This means **db-gen cannot detect nullable input parameters from PostgreSQL metadata**. This is a PostgreSQL limitation - the `information_schema.parameters` view doesn't provide reliable nullability information for function parameters.

**Workaround:** Use config overrides in `db-gen.json`:
```json
{
  "Generate": [
    {
      "Schema": "public",
      "Functions": {
        "your_function_name": {
          "Parameters": {
            "_your_parameter": {
              "IsNullable": true
            }
          }
        }
      }
    }
  ]
}
```

### Return Type Nullability

Return type nullability **IS** detected correctly from:
- `information_schema.columns.is_nullable` for TABLE returns
- `information_schema.attributes.is_nullable` for composite type returns

If a return column is not nullable in the database but you need it nullable in your code, use:
```json
{
  "Functions": {
    "your_function_name": {
      "Model": {
        "__column_name": {
          "IsNullable": true
        }
      }
    }
  }
}
```

## Testing These Queries

You can run these queries directly in psql or any PostgreSQL client:

### Get all functions in public schema:
```sql
-- Replace 'public' with your schema name
SELECT * FROM (...query 1...) WHERE r.specific_schema = 'public';
```

### Get parameters for a specific function:
```sql
-- First find the specific_name:
SELECT specific_name, type_udt_name, type_udt_schema
FROM information_schema.routines
WHERE routine_name = 'get_sendout_text'
  AND routine_schema = 'public';

-- Then use it in query 2:
-- Replace values with the results from above
SELECT * FROM (...query 2...)
WHERE specific_schema = 'public'
  AND specific_name = 'get_sendout_text_557992';
```

## Related PostgreSQL Documentation

- [information_schema.routines](https://www.postgresql.org/docs/current/infoschema-routines.html)
- [information_schema.parameters](https://www.postgresql.org/docs/current/infoschema-parameters.html)
- [information_schema.columns](https://www.postgresql.org/docs/current/infoschema-columns.html)
- [information_schema.attributes](https://www.postgresql.org/docs/current/infoschema-attributes.html)
