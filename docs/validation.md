# Validation System Guide

## Overview

The validation system allows you to define reusable validation rules that are automatically applied to database function parameters. It supports multiple validation strategies (FluentValidation, Manual, DataAnnotations, Assert) and can collect all errors or fail-fast.

## Key Features

✅ **Reusable Validation Rules** - Define once, apply to many parameters
✅ **Custom Executors** - Point to existing functions or generate inline code
✅ **Multi-Error Collection** - Return all validation errors at once or fail-fast
✅ **Strategy Agnostic** - Same rules work for FluentValidation, Manual, etc.
✅ **Language Agnostic** - Templates control output syntax
✅ **Function-Specific Overrides** - Global rules + per-function customization

## Configuration Structure

### Basic Setup

```json
{
  "Validation": {
    "ValidationStrategy": "Manual",
    "ValidationLocations": ["Controller", "Provider"],
    "ValidationBehavior": {
      "CollectAllErrors": true,
      "ReturnType": "ValidationResult"
    }
  }
}
```

### Validation Rule Definitions

Define validation rules that can be reused across parameters:

```json
{
  "ValidationRuleDefinitions": [
    {
      "Name": "CheckBusinessCode",
      "Type": "Custom",
      "ExecutorType": "ExistingFunction",
      "ExecutorReference": "ValidationHelpers.ValidateBusinessCode",
      "ErrorMessage": "Invalid business unit code",
      "ErrorCode": "INVALID_BUSINESS_CODE",
      "Templates": {
        "FluentValidation": ".Must(ValidationHelpers.ValidateBusinessCode).WithMessage(\"Invalid business unit code\").WithErrorCode(\"INVALID_BUSINESS_CODE\")",
        "Manual": {
          "Check": "ValidationHelpers.ValidateBusinessCode({{.ParamName}})",
          "ErrorBuilder": "new ValidationError(\"{{.ParamName}}\", \"Invalid business unit code\", \"INVALID_BUSINESS_CODE\")"
        }
      }
    }
  ]
}
```

### Parameter Validation Mappings

Map validation rules to parameter names globally:

```json
{
  "ParameterValidationMappings": [
    {
      "ParameterNames": ["businessunitcode", "businesscode", "businessunit"],
      "Rules": ["CheckBusinessCode"]
    },
    {
      "ParameterNames": ["email", "emailaddress"],
      "Rules": [
        "Required",
        "EmailFormat",
        {"Name": "MaxLength", "Value": 255}
      ]
    },
    {
      "ParameterNames": ["userid", "customerid"],
      "Rules": ["Required", "PositiveNumber"]
    }
  ]
}
```

### Function-Specific Validations

Override or extend global rules for specific functions:

```json
{
  "FunctionSpecificValidations": {
    "CreateUser": {
      "age": [
        {"Name": "Range", "Min": 18, "Max": 120}
      ]
    },
    "UpdateUserEmail": {
      "newemail": ["Required", "EmailFormat"]
    }
  }
}
```

## Validation Rule Types

### Built-in Rules

**Required:**
```json
{
  "Name": "Required",
  "Type": "Built-in",
  "ErrorCode": "REQUIRED_FIELD",
  "Templates": {
    "FluentValidation": ".NotNull().NotEmpty().WithErrorCode(\"REQUIRED_FIELD\")",
    "Manual": {
      "Check": "{{.ParamName}} != null && !string.IsNullOrEmpty({{.ParamName}})",
      "ErrorBuilder": "new ValidationError(\"{{.ParamName}}\", \"{{.ParamDisplayName}} is required\", \"REQUIRED_FIELD\")"
    }
  }
}
```

**Range:**
```json
{
  "Name": "Range",
  "Type": "Built-in",
  "ErrorCode": "OUT_OF_RANGE",
  "Templates": {
    "FluentValidation": ".InclusiveBetween({{.Min}}, {{.Max}}).WithErrorCode(\"OUT_OF_RANGE\")",
    "Manual": {
      "Check": "{{.ParamName}} >= {{.Min}} && {{.ParamName}} <= {{.Max}}",
      "ErrorBuilder": "new ValidationError(\"{{.ParamName}}\", \"Must be between {{.Min}} and {{.Max}}\", \"OUT_OF_RANGE\")"
    }
  }
}
```

### Custom Rules with Existing Functions

Point to existing validation functions in your codebase:

```json
{
  "Name": "CheckBusinessCode",
  "Type": "Custom",
  "ExecutorType": "ExistingFunction",
  "ExecutorReference": "ValidationHelpers.ValidateBusinessCode",
  "Templates": {
    "Manual": {
      "Check": "ValidationHelpers.ValidateBusinessCode({{.ParamName}})",
      "ErrorBuilder": "new ValidationError(\"{{.ParamName}}\", \"Invalid code\", \"INVALID_CODE\")"
    }
  }
}
```

### Custom Rules with Generated Code

Generate inline validation code:

```json
{
  "Name": "CheckTenantAccess",
  "Type": "Custom",
  "ExecutorType": "GenerateCode",
  "GeneratedCode": "tenantId > 0 && tenantId <= 1000",
  "Templates": {
    "Manual": {
      "Check": "{{.ParamName}} > 0 && {{.ParamName}} <= 1000",
      "ErrorBuilder": "new ValidationError(\"{{.ParamName}}\", \"Invalid tenant\", \"INVALID_TENANT\")"
    }
  }
}
```

## Validation Strategies

### FluentValidation

Generates validator classes using FluentValidation library:

```csharp
public class CreateUserValidator : AbstractValidator<CreateUserRequest>
{
    public CreateUserValidator()
    {
        RuleFor(x => x.Email)
            .NotNull().NotEmpty().WithErrorCode("REQUIRED_FIELD")
            .EmailAddress().WithErrorCode("INVALID_EMAIL")
            .MaximumLength(255).WithErrorCode("MAX_LENGTH_EXCEEDED");

        RuleFor(x => x.BusinessCode)
            .Must(ValidationHelpers.ValidateBusinessCode)
            .WithMessage("Invalid business unit code")
            .WithErrorCode("INVALID_BUSINESS_CODE");
    }
}
```

### Manual Validation

Generates inline validation in controllers:

**Collect All Errors:**
```csharp
[HttpPost]
public async Task<IActionResult> CreateUser(string email, string businessCode)
{
    var validationErrors = new List<ValidationError>();

    if (email == null || string.IsNullOrEmpty(email))
    {
        validationErrors.Add(new ValidationError("email", "Email is required", "REQUIRED_FIELD"));
    }

    if (!ValidationHelpers.ValidateBusinessCode(businessCode))
    {
        validationErrors.Add(new ValidationError("businessCode", "Invalid business unit code", "INVALID_BUSINESS_CODE"));
    }

    if (validationErrors.Any())
    {
        return BadRequest(new ValidationResult
        {
            Success = false,
            Errors = validationErrors
        });
    }

    // ... proceed with logic
}
```

**Fail-Fast:**
```csharp
[HttpPost]
public async Task<IActionResult> CreateUser(string email, string businessCode)
{
    if (email == null || string.IsNullOrEmpty(email))
    {
        return BadRequest(new ValidationError("email", "Email is required", "REQUIRED_FIELD"));
    }

    if (!ValidationHelpers.ValidateBusinessCode(businessCode))
    {
        return BadRequest(new ValidationError("businessCode", "Invalid code", "INVALID_BUSINESS_CODE"));
    }

    // ... proceed with logic
}
```

## Validation Behavior Options

### CollectAllErrors

**true** - Collect all validation errors before returning:
```json
{
  "success": false,
  "errors": [
    {
      "field": "email",
      "message": "Email is required",
      "code": "REQUIRED_FIELD"
    },
    {
      "field": "businessCode",
      "message": "Invalid business unit code",
      "code": "INVALID_BUSINESS_CODE"
    }
  ]
}
```

**false** - Fail on first validation error (fail-fast)

### ReturnType

**ValidationResult** - Simple error array:
```json
{
  "success": false,
  "errors": [
    {"field": "email", "message": "...", "code": "..."}
  ]
}
```

**ProblemDetails** - RFC 7807 format:
```json
{
  "type": "https://tools.ietf.org/html/rfc7807",
  "title": "Validation failed",
  "status": 400,
  "errors": {
    "email": ["Email is required"],
    "businessCode": ["Invalid business unit code"]
  }
}
```

## Template Placeholders

Available in validation templates:

- `{{.ParamName}}` - Parameter name
- `{{.ParamDisplayName}}` - Display name for parameter
- `{{.Min}}` - Minimum value (for Range)
- `{{.Max}}` - Maximum value (for Range)
- `{{.Value}}` - Generic value (for MaxLength, etc.)

## Example Use Cases

### Scenario 1: Business Code Used in 20 Functions

**Problem:** You have a `businessUnitCode` parameter in 20 functions and need consistent validation.

**Solution:**
```json
{
  "ValidationRuleDefinitions": [
    {
      "Name": "CheckBusinessCode",
      "ExecutorReference": "ValidationHelpers.ValidateBusinessCode",
      ...
    }
  ],
  "ParameterValidationMappings": [
    {
      "ParameterNames": ["businessunitcode", "businesscode"],
      "Rules": ["CheckBusinessCode"]
    }
  ]
}
```

Now all 20 functions automatically get the validation!

### Scenario 2: Email Validation with Length Limit

```json
{
  "ParameterValidationMappings": [
    {
      "ParameterNames": ["email", "emailaddress"],
      "Rules": [
        "Required",
        "EmailFormat",
        {"Name": "MaxLength", "Value": 255}
      ]
    }
  ]
}
```

### Scenario 3: Age Validation for Specific Function

```json
{
  "FunctionSpecificValidations": {
    "CreateUser": {
      "age": [
        {"Name": "Range", "Min": 18, "Max": 120}
      ]
    }
  }
}
```

## Integration with Context Mapping

Validation works seamlessly with context parameter mapping:

```json
{
  "UseUserContext": true,
  "ContextParameterMappings": [...],
  "Validation": {
    "ParameterValidationMappings": [
      {
        "ParameterNames": ["userid"],
        "Rules": ["Required", "PositiveNumber"]
      }
    ]
  }
}
```

Context parameters are automatically validated before being extracted from context.

## Benefits

1. **DRY Principle** - Define `CheckBusinessCode` once, use everywhere
2. **Consistency** - Same validation logic across all endpoints
3. **Type Safety** - Validation rules match database parameter types
4. **Flexibility** - Mix built-in rules with custom validators
5. **Documentation** - Error codes and messages serve as API documentation
6. **Testing** - Easy to test validation logic separately
7. **Migration** - Gradually add validations without breaking existing code

## Migration Strategy

1. **Start Small** - Add validation for critical parameters first
2. **Use Global Mappings** - Define rules for common parameters (email, userid)
3. **Add Custom Rules** - Create custom validators for business logic
4. **Enable Multi-Error** - Set `CollectAllErrors: true` for better UX
5. **Generate Controllers** - Use validation-enabled controller templates
6. **Test & Iterate** - Validate that error responses match expectations
