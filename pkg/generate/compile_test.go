package generate_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kalo-build/go-util/assertfile"
	"github.com/kalo-build/plugin-openapi-sdk-ts/pkg/generate"
	"github.com/stretchr/testify/suite"
)

type CompileTestSuite struct {
	assertfile.FileSuite

	TestDataPath    string
	GroundTruthPath string
}

func TestCompileTestSuite(t *testing.T) {
	suite.Run(t, new(CompileTestSuite))
}

func (suite *CompileTestSuite) SetupTest() {
	_, filename, _, _ := runtime.Caller(0)
	suite.TestDataPath = filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
	suite.GroundTruthPath = filepath.Join(suite.TestDataPath, "ground-truth")
}

func (suite *CompileTestSuite) TestGenerateSDK() {
	inputDir := filepath.Join(suite.TestDataPath, "input")
	outputDir := suite.T().TempDir()

	cfg := generate.Config{
		InputDir:       inputDir,
		OutputDir:      outputDir,
		SpecFileName:   "openapi.yaml",
		GenerateTypes:  true,
		GenerateClient: true,
	}
	cfg.Resolve()

	err := generate.GenerateSDK(cfg)
	suite.NoError(err)

	typesPath := filepath.Join(outputDir, "types.ts")
	gtTypesPath := filepath.Join(suite.GroundTruthPath, "types.ts")
	suite.FileExists(typesPath)
	suite.FileEquals(typesPath, gtTypesPath)

	clientPath := filepath.Join(outputDir, "client.ts")
	gtClientPath := filepath.Join(suite.GroundTruthPath, "client.ts")
	suite.FileExists(clientPath)
	suite.FileEquals(clientPath, gtClientPath)
}

func (suite *CompileTestSuite) TestGenerateSDK_TypesOnly() {
	inputDir := filepath.Join(suite.TestDataPath, "input")
	outputDir := suite.T().TempDir()

	cfg := generate.Config{
		InputDir:       inputDir,
		OutputDir:      outputDir,
		SpecFileName:   "openapi.yaml",
		GenerateTypes:  true,
		GenerateClient: false,
	}
	cfg.Resolve()

	err := generate.GenerateSDK(cfg)
	suite.NoError(err)

	typesPath := filepath.Join(outputDir, "types.ts")
	gtTypesPath := filepath.Join(suite.GroundTruthPath, "types.ts")
	suite.FileExists(typesPath)
	suite.FileEquals(typesPath, gtTypesPath)

	clientPath := filepath.Join(outputDir, "client.ts")
	_, err = os.Stat(clientPath)
	suite.True(os.IsNotExist(err), "client.ts should not exist when GenerateClient is false")
}
