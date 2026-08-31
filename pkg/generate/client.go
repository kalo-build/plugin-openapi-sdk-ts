package generate

import (
	"fmt"
	"strings"
)

// GenerateClient produces a typed TypeScript fetch client from the document's paths.
func GenerateClient(doc *Document, typesImportPath string) string {
	var b strings.Builder

	writeClientImports(&b, doc, typesImportPath)
	b.WriteString("\n")
	writeApiError(&b)
	b.WriteString("\n")
	writeClientClass(&b, doc)

	return b.String()
}

func writeClientImports(b *strings.Builder, doc *Document, typesImportPath string) {
	typeNames := collectReferencedTypes(doc)
	if len(typeNames) == 0 {
		return
	}
	b.WriteString(fmt.Sprintf("import type { %s } from \"%s\";\n", strings.Join(typeNames, ", "), typesImportPath))
}

func collectReferencedTypes(doc *Document) []string {
	seen := map[string]bool{}
	for _, pe := range doc.Paths {
		for _, op := range pe.Operations {
			if op.RequestBody != nil && op.RequestBody.Ref != "" {
				seen[tsTypeName(refToTypeName(op.RequestBody.Ref))] = true
			}
			for _, sr := range op.Responses {
				collectSchemaRefTypes(sr, seen)
			}
		}
	}
	var names []string
	for name := range seen {
		names = append(names, name)
	}
	sortStrings(names)
	return names
}

func collectSchemaRefTypes(sr *SchemaRef, seen map[string]bool) {
	if sr == nil {
		return
	}
	if sr.Ref != "" {
		seen[tsTypeName(refToTypeName(sr.Ref))] = true
		return
	}
	if sr.Schema == nil {
		return
	}
	for _, prop := range sr.Schema.Properties {
		if prop.Ref != "" {
			seen[tsTypeName(refToTypeName(prop.Ref))] = true
		}
		if prop.Schema.HasType("array") && prop.Schema.Items != nil && prop.Schema.Items.Ref != "" {
			seen[tsTypeName(refToTypeName(prop.Schema.Items.Ref))] = true
		}
	}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func writeApiError(b *strings.Builder) {
	b.WriteString(`export class ApiError extends Error {
  statusCode: number;
  body: string;

  constructor(statusCode: number, body: string) {
    super(` + "`API error ${statusCode}: ${body}`" + `);
    this.name = "ApiError";
    this.statusCode = statusCode;
    this.body = body;
  }
}
`)
}

func writeClientClass(b *strings.Builder, doc *Document) {
	b.WriteString(`export class Client {
  baseUrl: string;
  private fetchFn: typeof fetch;

  constructor(baseUrl: string, fetchFn: typeof fetch = fetch.bind(globalThis)) {
    this.baseUrl = baseUrl;
    this.fetchFn = fetchFn;
  }
`)

	for _, pe := range doc.Paths {
		for _, op := range pe.Operations {
			b.WriteString("\n")
			writeClientMethod(b, pe.Path, op)
		}
	}

	b.WriteString("}\n")
}

func writeClientMethod(b *strings.Builder, path string, op Operation) {
	funcName := operationFuncName(op, path)
	returnType := operationReturnType(op)
	hasBody := op.Method == "POST" || op.Method == "PUT" || op.Method == "PATCH"
	bodyType := requestBodyType(op)
	pathParams := extractPathParams(op)
	queryParams := extractQueryParams(op)

	sig := buildTsSignature(funcName, pathParams, queryParams, hasBody, bodyType, returnType)
	b.WriteString(sig)
	b.WriteString(" {\n")

	writeUrlBuilder(b, path, pathParams, queryParams)

	method := strings.ToUpper(op.Method)

	if hasBody && bodyType != "" {
		b.WriteString(fmt.Sprintf("    const resp = await this.fetchFn(url, {\n"))
		b.WriteString(fmt.Sprintf("      method: \"%s\",\n", method))
		b.WriteString("      headers: { \"Content-Type\": \"application/json\", \"Accept\": \"application/json\" },\n")
		b.WriteString("      body: JSON.stringify(body),\n")
		b.WriteString("    });\n")
	} else {
		b.WriteString(fmt.Sprintf("    const resp = await this.fetchFn(url, {\n"))
		b.WriteString(fmt.Sprintf("      method: \"%s\",\n", method))
		b.WriteString("      headers: { \"Accept\": \"application/json\" },\n")
		b.WriteString("    });\n")
	}

	b.WriteString("    if (!resp.ok) {\n")
	b.WriteString("      const body = await resp.text();\n")
	b.WriteString("      throw new ApiError(resp.status, body);\n")
	b.WriteString("    }\n")

	if returnType != "" {
		b.WriteString(fmt.Sprintf("    return resp.json() as Promise<%s>;\n", returnType))
	}

	b.WriteString("  }\n")
}

func buildTsSignature(funcName string, pathParams, queryParams []tsParamInfo, hasBody bool, bodyType, returnType string) string {
	var params []string
	for _, p := range pathParams {
		params = append(params, fmt.Sprintf("%s: %s", p.tsName, p.tsType))
	}
	for _, p := range queryParams {
		params = append(params, fmt.Sprintf("%s?: %s", p.tsName, p.tsType))
	}
	if hasBody && bodyType != "" {
		params = append(params, fmt.Sprintf("body: %s", bodyType))
	}

	ret := "Promise<void>"
	if returnType != "" {
		ret = fmt.Sprintf("Promise<%s>", returnType)
	}

	return fmt.Sprintf("  async %s(%s): %s", funcName, strings.Join(params, ", "), ret)
}

func writeUrlBuilder(b *strings.Builder, path string, pathParams []tsParamInfo, queryParams []tsParamInfo) {
	if len(pathParams) > 0 {
		tmpl := path
		for _, p := range pathParams {
			placeholder := "{" + p.name + "}"
			tmpl = strings.Replace(tmpl, placeholder, "${encodeURIComponent(String("+p.tsName+"))}", 1)
		}
		b.WriteString(fmt.Sprintf("    let url = `${this.baseUrl}%s`;\n", tmpl))
	} else {
		b.WriteString(fmt.Sprintf("    let url = `${this.baseUrl}%s`;\n", path))
	}

	if len(queryParams) > 0 {
		b.WriteString("    const query = new URLSearchParams();\n")
		for _, p := range queryParams {
			b.WriteString(fmt.Sprintf("    if (%s !== undefined) query.set(\"%s\", %s);\n", p.tsName, p.name, p.tsName))
		}
		b.WriteString("    if (query.toString()) url += `?${query.toString()}`;\n")
	}
}

type tsParamInfo struct {
	name   string
	tsName string
	tsType string
}

func extractPathParams(op Operation) []tsParamInfo {
	var params []tsParamInfo
	for _, p := range op.Parameters {
		if p.In == "path" {
			tsType := "string"
			if p.Schema != nil {
				tsType = schemaToTsType(*p.Schema)
			}
			params = append(params, tsParamInfo{
				name:   p.Name,
				tsName: toCamelCase(p.Name),
				tsType: tsType,
			})
		}
	}
	return params
}

func extractQueryParams(op Operation) []tsParamInfo {
	var params []tsParamInfo
	for _, p := range op.Parameters {
		if p.In == "query" {
			params = append(params, tsParamInfo{
				name:   p.Name,
				tsName: toCamelCase(p.Name),
				tsType: "string",
			})
		}
	}
	return params
}

func operationFuncName(op Operation, path string) string {
	if op.OperationID != "" {
		return toCamelCase(op.OperationID)
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	resource := ""
	if len(parts) > 0 {
		resource = toPascalCase(parts[len(parts)-1])
	}
	return toCamelCase(methodPrefix(op.Method) + resource)
}

func methodPrefix(method string) string {
	switch method {
	case "GET":
		return "Get"
	case "POST":
		return "Create"
	case "PUT", "PATCH":
		return "Update"
	case "DELETE":
		return "Delete"
	default:
		return method
	}
}

func operationReturnType(op Operation) string {
	for _, code := range []string{"200", "201"} {
		sr, ok := op.Responses[code]
		if !ok || sr == nil {
			continue
		}
		if sr.Ref != "" {
			return tsTypeName(refToTypeName(sr.Ref))
		}
		if sr.Schema != nil {
			s := sr.Schema
			if s.HasType("object") && len(s.Properties) > 0 {
				for _, prop := range s.Properties {
					if prop.Name == "data" {
						if prop.Schema.HasType("array") && prop.Schema.Items != nil {
							if prop.Schema.Items.Ref != "" {
								return tsTypeName(refToTypeName(prop.Schema.Items.Ref)) + "[]"
							}
						}
						if prop.Ref != "" {
							return tsTypeName(refToTypeName(prop.Ref))
						}
					}
				}
			}
		}
	}
	return ""
}

func requestBodyType(op Operation) string {
	if op.RequestBody == nil {
		return ""
	}
	if op.RequestBody.Ref != "" {
		return tsTypeName(refToTypeName(op.RequestBody.Ref))
	}
	return ""
}
