# Example templates

Real-world-shaped, **anonymized** db-gen templates you can copy and adapt. They
come from a production project with the branding stripped — replace the
`MyApp.*` namespaces (and the `SecureLogging` helper) with your own.

These are illustrative starting points, not a supported API. The authoritative,
minimal templates used by the test suite live in [`test/templates/`](../test/templates)
and [`test/e2e/templates/`](../test/e2e/templates); the data model each template
receives is documented in [docs/templating.md](../docs/templating.md).

## Layout

| File | Kind | Target |
|------|------|--------|
| [`csharp/dbcontext.gotmpl`](./csharp/dbcontext.gotmpl) | `DbContextTemplate` | C# / Npgsql, `DotNext` `Optional<T>` params, jsonb handling |
| [`csharp/model.gotmpl`](./csharp/model.gotmpl) | `ModelTemplate` | C# POCO with `[DbColumnMapping]` |
| [`csharp/processor.gotmpl`](./csharp/processor.gotmpl) | `ProcessorTemplate` | C# reader → model mapper, null-aware |
| [`csharp/common-provider.gotmpl`](./csharp/common-provider.gotmpl) | additional generator (`single-file`) | C# provider with **secure logging** driven by `SecurityLevel` |
| [`typescript/model.gotmpl`](./typescript/model.gotmpl) | additional generator (`per-routine`) | TypeScript interface per return shape |

## Wiring them up

Point the core template keys at the C# files and register the extras as
[additional generators](../docs/context-mapping.md):

```json
{
  "DbContextTemplate": "./examples/csharp/dbcontext.gotmpl",
  "ModelTemplate": "./examples/csharp/model.gotmpl",
  "ProcessorTemplate": "./examples/csharp/processor.gotmpl",
  "GenerateModels": true,
  "GenerateProcessors": true,
  "AdditionalGenerators": [
    {
      "Name": "CommonProvider",
      "Enabled": true,
      "Template": "./examples/csharp/common-provider.gotmpl",
      "OutputFolder": "./generated/providers",
      "FileName": "CommonProvider.cs",
      "GenerationType": "single-file"
    },
    {
      "Name": "TypeScript",
      "Enabled": true,
      "Template": "./examples/typescript/model.gotmpl",
      "OutputFolder": "./generated/ts",
      "FileExtension": ".ts",
      "FileCase": "pascalcase",
      "GenerationType": "per-routine"
    }
  ]
}
```

> Set `FileCase` on per-routine generators — otherwise db-gen logs
> `unknown case, this should never happen` and falls back to the raw name.

## Secure logging (the `common-provider` example)

`common-provider.gotmpl` shows how to turn each parameter's
[`SecurityLevel`](../docs/templating.md#parameter-security-levels) into a log
line that never leaks secrets:

- `none` → logged as-is
- `secure` → plain text when `SecureLogging.Insecure` is `true`, else `******`
- `strict` → always `******`
- `omit` → dropped from the message **and** the argument list

`SecureLogging.Insecure` is the runtime flag the feature request calls
`db-gen:insecure-logging` — wire it to your configuration at startup. db-gen
only resolves the level; the template renders the masking, keeping the tool
language-agnostic. Assign levels in config via `ParameterSecurityMappings`
(global by name), per-function `Parameters[].SecurityLevel`, and
`DefaultParameterSecurityLevel` — see
[docs/configuration.md → parameter security](../docs/configuration.md#parameter-security).
