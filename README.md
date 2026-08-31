# plugin-openapi-sdk-ts

Kalo plugin that generates a typed TypeScript HTTP client SDK from an OpenAPI 3.1 specification using [libopenapi](https://github.com/pb33f/libopenapi). Mirrors `plugin-openapi-sdk-go` for the TypeScript ecosystem.

## What it generates

| File | Contents |
|------|----------|
| `types.ts` | TypeScript interfaces from component schemas |
| `client.ts` | Typed `fetch` client with async methods per operation |

### Generated types example

```typescript
export interface Company {
  created_at: string;
  id: string;
  name: string;
  updated_at: string;
}

export interface CompanyCreate {
  id?: string;
  name: string;
}
```

### Generated client example

```typescript
import { Client } from "./client";

const client = new Client("https://api.example.com");

const companies = await client.listCompany();
const company = await client.createCompany({ name: "Acme" });
const fetched = await client.getCompany("uuid-here");
const updated = await client.updateCompany("uuid-here", { name: "Acme Inc" });
await client.deleteCompany("uuid-here");
```

The client accepts an optional `fetch` implementation for testing or custom transports:

```typescript
const client = new Client("https://api.example.com", customFetch);
```

## Input / output

| Direction | Format | Description |
|-----------|--------|-------------|
| **Input** | `KA:OA1:YAML1` | OpenAPI 3.1 YAML specification |
| **Output** | `KA:OA1:TS_CLIENT1` | Generated TypeScript HTTP client source files |

## Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `specFileName` | string | `"openapi.yaml"` | Name of the spec file to read from input |
| `generateTypes` | boolean | `true` | Generate type definitions (`types.ts`). Set `false` when using external types from another plugin. |
| `generateClient` | boolean | `true` | Generate HTTP client (`client.ts`) |
| `typesImportPath` | string | `"./types"` | Import path the client uses for type imports. Set to an external path when types come from another plugin. |

### Using with external types (e.g. plugin-morphe-ts-types)

When your pipeline already generates TypeScript types from Morphe (via `plugin-morphe-ts-types`), you can skip duplicate type generation and point the client at the existing types:

```yaml
plugins:
  "@kalo-build/plugin-openapi-sdk-ts":
    input:
      store: KA_OA_YAML
    output:
      format: "KA:OA1:TS_CLIENT1"
      store: KA_OA_TS_CLIENT
    config:
      generateTypes: false
      typesImportPath: "@myapp/generated/types"
```

This follows the same pattern as `plugin-morphe-ts-zod-bridge` for cross-plugin type references.

## Type mapping

| OpenAPI Type | Format | TypeScript Type |
|-------------|--------|-----------------|
| `string` | — | `string` |
| `string` | `date-time` | `string` |
| `string` | `uuid` | `string` |
| `integer` | — | `number` |
| `integer` | `int64` | `number` |
| `integer` | `int32` | `number` |
| `number` | — | `number` |
| `number` | `float` | `number` |
| `boolean` | — | `boolean` |
| `array` | — | `T[]` |
| `$ref` | — | Referenced type name |

Optional fields (not in `required`) use `?:` syntax.

## Usage in `kalo.yaml`

### Standalone (generates its own types)

```yaml
stores:
  KA_OA_YAML:
    type: localFileSystem
    path: ./docs/openapi
  KA_OA_TS_CLIENT:
    type: localFileSystem
    path: ./generated/client

plugins:
  "@kalo-build/plugin-openapi-sdk-ts":
    input:
      store: KA_OA_YAML
    output:
      format: "KA:OA1:TS_CLIENT1"
      store: KA_OA_TS_CLIENT
```

### With Morphe pipeline (external types)

```yaml
stores:
  KA_OA_YAML:
    type: localFileSystem
    path: ./docs/openapi
  KA_MO_TS:
    type: localFileSystem
    path: ./generated/types
  KA_OA_TS_CLIENT:
    type: localFileSystem
    path: ./generated/client

plugins:
  "@kalo-build/plugin-morphe-ts-types":
    input:
      store: KA_MO_YAML
    output:
      format: "KA:MO1:TS1"
      store: KA_MO_TS

  "@kalo-build/plugin-openapi-sdk-ts":
    input:
      store: KA_OA_YAML
    output:
      format: "KA:OA1:TS_CLIENT1"
      store: KA_OA_TS_CLIENT
    config:
      generateTypes: false
      typesImportPath: "@myapp/generated/types"
```

## How it works

1. Reads the OpenAPI YAML spec from the input directory
2. Parses and builds a V3 model using `pb33f/libopenapi`
3. Converts to an intermediate representation (decoupling parser from generators)
4. Generates TypeScript interfaces from `components.schemas`
5. Generates a typed `fetch` client from `paths` and their operations
6. Writes output files to the output directory

## Project structure

```
plugin-openapi-sdk-ts/
├── cmd/plugin/main.go            # WASM entrypoint
├── pkg/generate/
│   ├── config.go                 # Configuration with defaults
│   ├── spec.go                   # Intermediate representation types
│   ├── generate.go               # Orchestrator (parse + convert + generate + write)
│   ├── generate_test.go          # Unit tests
│   ├── compile_test.go           # Ground-truth integration tests
│   ├── types.go                  # TypeScript interface generation from schemas
│   ├── client.go                 # Fetch client generation from paths
│   └── util.go                   # Naming helpers
├── testdata/
│   ├── input/openapi.yaml        # Test fixture (OpenAPI spec)
│   └── ground-truth/             # Expected output for integration tests
│       ├── types.ts
│       └── client.ts
├── plugin.yaml                   # Kalo plugin manifest
└── dist/                         # WASM output (gitignored)
```

## Building

```bash
# Native binary
go build ./cmd/plugin

# WASM (for Kalo CLI)
GOOS=wasip1 GOARCH=wasm go build -o dist/plugin.wasm cmd/plugin/main.go
```

## Testing

```bash
go test ./...
```
