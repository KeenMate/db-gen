# Context Parameter Mapping & Additional Generators Guide

## Overview

This guide covers two major features added to db-gen:
1. **Context Parameter Mappings** - Automatically map database function parameters to user context properties
2. **Additional Generators** - Extensible system for generating additional code outputs (TypeScript, Providers, Controllers, etc.)

## Context Parameter Mappings

### What Problem Does This Solve?

Many database functions require authentication/context parameters like `userid`, `tenantid`, `createdby`, etc. Instead of passing these as explicit parameters in your application code, you can map them to a UserContext object.

### Configuration

```json
{
  "UseUserContext": true,
  "UserContextParameterName": "ctx",
  "UserContextType": "UserContext",
  "ContextParameterMappings": [
    {
      "ParameterNames": ["userid"],
      "ContextPath": "User.UserId"
    },
    {
      "ParameterNames": ["displaylanguagecode"],
      "ContextPath": "Locale"
    },
    {
      "ParameterNames": ["createdby", "updatedby", "deletedby", "modifiedby", "removedby"],
      "ContextPath": "Username"
    },
    {
      "ParameterNames": ["tenantid"],
      "ContextPath": "SelectedTenant.TenantId"
    }
  ]
}
```

### How It Works

**Without Context (UseUserContext: false):**
```csharp
public async Task<List<User>> GetUsersAsync(int userId, int tenantId, string searchTerm)
{
    return await dbContext.GetUsersAsync(userId, tenantId, searchTerm);
}
```

**With Context (UseUserContext: true):**
```csharp
public async Task<List<User>> GetUsersAsync(UserContext ctx, string searchTerm)
{
    return await dbContext.GetUsersAsync(ctx.User.UserId, ctx.SelectedTenant.TenantId, searchTerm);
}
```

### Template Access

Templates can access context information through the `Routine` object:

```gotmpl
{{range $routine := .Functions}}
  {{- if $routine.UsesUserContext}}
    // This routine uses user context
    {{range $param := $routine.ContextParameters}}
      // Context param: {{$param.PropertyName}} -> {{$param.ContextPath}}
    {{end}}
    {{range $param := $routine.RegularParameters}}
      // Regular param: {{$param.PropertyName}}
    {{end}}
  {{- end}}
{{end}}
```

### Language Agnostic

This feature is **completely language agnostic**. The configuration provides data, and templates decide how to use it:

**C# Example:**
```gotmpl
{{$.Config.UserContextParameterName}}.{{$param.ContextPath}}
// Produces: ctx.User.UserId
```

**Elixir Example:**
```gotmpl
%{{$.Config.UserContextType}}{} = {{$.Config.UserContextParameterName}}
{{$.Config.UserContextParameterName}}.{{$param.ContextPath}}
// Produces: %UserContext{} = ctx
// ctx.user.user_id
```

## Additional Generators

### What Problem Does This Solve?

Previously, you needed a separate tool (like csharp-model-generator) to generate TypeScript models, Providers, or other code. Now, db-gen can generate everything in one pass using its internal data structures.

### Configuration

```json
{
  "AdditionalGenerators": [
    {
      "Name": "TypeScript",
      "Enabled": true,
      "Template": "./templates/typescript.gotmpl",
      "OutputFolder": "./output/typescript",
      "FileExtension": ".ts",
      "FileCase": "camelcase",
      "GenerationType": "per-routine"
    },
    {
      "Name": "Provider",
      "Enabled": true,
      "Template": "./templates/provider.gotmpl",
      "OutputFolder": "./output",
      "FileName": "CommonProvider.cs",
      "GenerationType": "single-file"
    },
    {
      "Name": "Controller",
      "Enabled": false,
      "Template": "./templates/controller.gotmpl",
      "OutputFolder": "./output/controllers",
      "FileExtension": ".cs",
      "FileCase": "pascalcase",
      "GenerationType": "per-routine"
    }
  ]
}
```

### Generator Types

#### per-routine
Generates one file for each database function/procedure.

- Uses `ModelTemplateData` structure
- Has access to single `Routine` object
- Output filename based on routine name + `FileExtension`
- Example: TypeScript interfaces, individual controllers

#### single-file
Generates one file containing all routines.

- Uses `DbContextData` structure
- Has access to all `Functions` array
- Output filename is `FileName`
- Example: CommonProvider, DbContext, SignalR Hub

### Configuration Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `Name` | string | Yes | Display name for logging |
| `Enabled` | bool | Yes | Enable/disable this generator |
| `Template` | string | Yes | Path to .gotmpl template file |
| `OutputFolder` | string | Yes | Where to write generated files |
| `FileName` | string | Conditional | Required for single-file generators |
| `FileExtension` | string | Conditional | Required for per-routine generators |
| `FileCase` | string | Optional | "snakecase", "camelcase", or "pascalcase" |
| `GenerationType` | string | Optional | Auto-detected: "per-routine" or "single-file" |

### Example Templates

#### TypeScript (per-routine)
```gotmpl
export interface {{.Routine.ModelName}} {
{{- range $property := .Routine.ReturnProperties}}
  {{camelCased $property.PropertyName}}: {{if eq $property.PropertyType "int"}}number{{else}}any{{end}};
{{- end}}
}
```

#### Provider (single-file)
```gotmpl
public class CommonProvider
{
{{range $routine := .Functions}}
  public async Task<List<{{$routine.ModelName}}>> {{$routine.FunctionName}}Async(
    {{if $routine.UsesUserContext}}{{$.Config.UserContextType}} {{$.Config.UserContextParameterName}}{{end}}
    {{range $param := $routine.RegularParameters}}, {{$param.PropertyType}} {{$param.PropertyName}}{{end}}
  )
  {
    return await dbContext.{{$routine.FunctionName}}Async(
      {{range $param := $routine.Parameters}}
        {{if $param.IsContextParameter}}
          {{$.Config.UserContextParameterName}}.{{$param.ContextPath}}
        {{else}}
          {{$param.PropertyName}}
        {{end}},
      {{end}}
    );
  }
{{end}}
}
```

## Future Use Cases

The flexible generator system enables:

- **C# Controllers** - Auto-generate ASP.NET Core API endpoints
- **SignalR Hubs** - Generate hub methods for real-time communication
- **Elixir Channels** - Generate Phoenix Channel handlers
- **GraphQL Schemas** - Generate schema definitions and resolvers
- **OpenAPI/Swagger** - Generate API documentation
- **Test Fixtures** - Generate test data and mocks
- **Client SDKs** - Generate client libraries in any language

## Benefits

1. **Single Source of Truth** - Database functions drive all code generation
2. **No Re-parsing** - Uses db-gen's internal data structures directly
3. **Consistency** - All outputs guaranteed to match database schema
4. **Efficiency** - One database query, multiple outputs
5. **Extensibility** - Add new generators without code changes
6. **Type Safety** - Compile-time checking across all layers

## Migration from csharp-model-generator

If you're currently using `csharp-model-generator`:

1. Add context mappings to `db-gen.json`
2. Add TypeScript and Provider generators to `AdditionalGenerators`
3. Copy/adapt your template logic to Go templates
4. Run `db-gen generate` once instead of two tools
5. Remove `csharp-model-generator` from your pipeline
