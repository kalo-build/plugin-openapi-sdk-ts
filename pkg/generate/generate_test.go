package generate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func TestGenerateSDK_Full(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	outputDir := t.TempDir()

	cfg := Config{
		InputDir:       inputDir,
		OutputDir:      outputDir,
		SpecFileName:   "openapi.yaml",
		GenerateTypes:  true,
		GenerateClient: true,
	}
	cfg.Resolve()

	err := GenerateSDK(cfg)
	require.NoError(t, err)

	typesContent, err := os.ReadFile(filepath.Join(outputDir, "types.ts"))
	require.NoError(t, err)
	typesStr := string(typesContent)

	assert.Contains(t, typesStr, "export interface Company")
	assert.Contains(t, typesStr, "export interface CompanyCreate")
	assert.Contains(t, typesStr, "export interface CompanyUpdate")
	assert.Contains(t, typesStr, "name: string")
	assert.Contains(t, typesStr, "created_at: string")

	clientContent, err := os.ReadFile(filepath.Join(outputDir, "client.ts"))
	require.NoError(t, err)
	clientStr := string(clientContent)

	assert.Contains(t, clientStr, "export class Client")
	assert.Contains(t, clientStr, "listCompany")
	assert.Contains(t, clientStr, "createCompany")
	assert.Contains(t, clientStr, "getCompany")
	assert.Contains(t, clientStr, "updateCompany")
	assert.Contains(t, clientStr, "deleteCompany")
	assert.Contains(t, clientStr, "from \"./types\"")
}

func TestGenerateSDK_TypesOnly(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	outputDir := t.TempDir()

	cfg := Config{
		InputDir:       inputDir,
		OutputDir:      outputDir,
		SpecFileName:   "openapi.yaml",
		GenerateTypes:  true,
		GenerateClient: false,
	}
	cfg.Resolve()

	err := GenerateSDK(cfg)
	require.NoError(t, err)

	_, err = os.ReadFile(filepath.Join(outputDir, "types.ts"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "client.ts"))
	assert.True(t, os.IsNotExist(err))
}

func TestGenerateSDK_ClientOnly(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	outputDir := t.TempDir()

	cfg := Config{
		InputDir:       inputDir,
		OutputDir:      outputDir,
		SpecFileName:   "openapi.yaml",
		GenerateTypes:  false,
		GenerateClient: true,
	}
	cfg.Resolve()

	err := GenerateSDK(cfg)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "types.ts"))
	assert.True(t, os.IsNotExist(err))

	_, err = os.ReadFile(filepath.Join(outputDir, "client.ts"))
	require.NoError(t, err)
}

func TestGenerateSDK_ExternalTypes(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	outputDir := t.TempDir()

	cfg := Config{
		InputDir:        inputDir,
		OutputDir:       outputDir,
		SpecFileName:    "openapi.yaml",
		GenerateTypes:   false,
		GenerateClient:  true,
		TypesImportPath: "@myapp/generated/types",
	}
	cfg.Resolve()

	err := GenerateSDK(cfg)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(outputDir, "types.ts"))
	assert.True(t, os.IsNotExist(err))

	clientContent, err := os.ReadFile(filepath.Join(outputDir, "client.ts"))
	require.NoError(t, err)
	clientStr := string(clientContent)

	assert.Contains(t, clientStr, "from \"@myapp/generated/types\"")
	assert.NotContains(t, clientStr, "from \"./types\"")
}

func TestParseSpec(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	specBytes, err := os.ReadFile(filepath.Join(inputDir, "openapi.yaml"))
	require.NoError(t, err)

	doc, err := ParseSpec(specBytes)
	require.NoError(t, err)
	assert.NotNil(t, doc)
	assert.Equal(t, "Test API", doc.Info.Title)
	assert.Greater(t, len(doc.Paths), 0)
	assert.Greater(t, len(doc.Schemas), 0)
}

func TestGenerateTypes_Schemas(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	specBytes, err := os.ReadFile(filepath.Join(inputDir, "openapi.yaml"))
	require.NoError(t, err)

	doc, err := ParseSpec(specBytes)
	require.NoError(t, err)

	types := GenerateTypes(doc)

	assert.Contains(t, types, "export interface Company")
	assert.Contains(t, types, "id: string")
	assert.Contains(t, types, "name: string")
	assert.Contains(t, types, "created_at: string")
	assert.Contains(t, types, "updated_at: string")
}

func TestGenerateClient_Operations(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	specBytes, err := os.ReadFile(filepath.Join(inputDir, "openapi.yaml"))
	require.NoError(t, err)

	doc, err := ParseSpec(specBytes)
	require.NoError(t, err)

	client := GenerateClient(doc, "./types")

	assert.Contains(t, client, "async listCompany()")
	assert.Contains(t, client, "async createCompany(body: CompanyCreate)")
	assert.Contains(t, client, "async getCompany(id: string)")
	assert.Contains(t, client, "async updateCompany(id: string, body: CompanyUpdate)")
	assert.Contains(t, client, "async deleteCompany(id: string)")
}

func TestConfig_Resolve(t *testing.T) {
	cfg := Config{}
	cfg.Resolve()
	assert.Equal(t, "openapi.yaml", cfg.SpecFileName)
	assert.Equal(t, "./types", cfg.TypesImportPath)
}

func TestConfig_Resolve_PreservesExplicit(t *testing.T) {
	cfg := Config{
		SpecFileName:    "api.yaml",
		TypesImportPath: "@shared/types",
	}
	cfg.Resolve()
	assert.Equal(t, "api.yaml", cfg.SpecFileName)
	assert.Equal(t, "@shared/types", cfg.TypesImportPath)
}
