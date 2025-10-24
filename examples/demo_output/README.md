# Demo Output: Comprehensive API Generation Examples

This example demonstrates all the different Go types and patterns supported by the `apigen` package. It generates demo output files that can be visually inspected to verify correct API description generation.

## Overview

The example showcases how various Go constructs are represented in the generated JSON API descriptions:

- **Basic types**: `string`, `int`, `bool`, `float64`
- **Nested structs**: Structs containing other structs
- **Slice types**: Arrays/slices of structs and basic types
- **Map types**: Maps with various key/value combinations
- **Complex nested structures**: Deep nesting with multiple levels
- **Pointer types**: Pointers to structs
- **External package types**: Types from other packages (e.g., `time.Time`)

## Running the Demo

```bash
cd examples/demo_output
go run main.go
```

This will:
1. Parse the `api` package containing comprehensive type examples
2. Apply different filtering strategies to demonstrate various use cases
3. Generate demo files in the `demo_outputs/` directory
4. Create individual method files for detailed inspection

## Generated Demo Files

The demo generates several categories of golden files:

### Filter-based files:
- `all_methods.json` - All methods in the API
- `handle_prefix.json` - Methods starting with "Handle"
- `process_prefix.json` - Methods starting with "Process"
- `update_prefix.json` - Methods starting with "Update"
- `schedule_prefix.json` - Methods starting with "Schedule"

### Individual method files:
- `method_HandleBasicTypes.json`
- `method_ProcessCompany.json`
- `method_ProcessUsers.json`
- `method_ProcessProfiles.json`
- `method_ProcessTeam.json`
- `method_UpdateConfig.json`
- `method_ScheduleEvent.json`

## Example Output Structure

Each golden file contains a JSON structure like:

```json
{
  "apiName": "DemoAPI",
  "methods": {
    "ProcessUsers": {
      "description": "processes a slice of users",
      "parameters": {
        "users": {
          "type": "[]User",
          "fields": {
            "id": {"type": "int"},
            "name": {"type": "string"},
            "email": {"type": "string"},
            "isActive": {"type": "bool"},
            "tags": {"type": "[]string"}
          }
        }
      }
    }
  }
}
```

## Inspecting Results

After running the demo, examine the generated JSON files to see how:

- Basic types are represented as simple strings
- Struct types include nested `fields` objects with type information
- Slice types use `[]ElementType` notation
- Map types use `map[KeyType]ValueType` notation
- Struct tags (like `json:"fieldName"`) are preserved in the `annotations` field
- Nested structures are fully resolved with complete field information

## Current Limitations

Some advanced features may not be fully implemented yet:
- Deep nested struct resolution (fields may be empty for complex nesting)
- Interface{} types
- Complex generic types
- Circular type references

The demo files serve as both documentation and examples for verifying these features.</content>
