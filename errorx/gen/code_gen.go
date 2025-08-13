package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for code generation
type Config struct {
	AppName        string
	AppCode        int
	ImportPath     string
	ScriptDir      string
	ProjectRoot    string
	OutputTemplate string
}

// loadYAML loads a YAML file and returns the data as map[string]interface{}
func loadYAML(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}

	return result, nil
}

// validateBusinessCodes validates that all business codes are unique within each app
func validateBusinessCodes(metadata map[string]interface{}) error {
	apps, ok := metadata["app"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid metadata format: app field not found or not an array")
	}

	for _, appInterface := range apps {
		app, ok := appInterface.(map[string]interface{})
		if !ok {
			continue
		}

		appName, _ := app["name"].(string)
		businessList, ok := app["business"].([]interface{})
		if !ok {
			continue
		}

		seenCodes := make(map[int]string)
		for _, bizInterface := range businessList {
			biz, ok := bizInterface.(map[string]interface{})
			if !ok {
				continue
			}

			bizName, _ := biz["name"].(string)
			bizCode, _ := biz["code"].(int)

			if existingBiz, exists := seenCodes[bizCode]; exists {
				return fmt.Errorf("duplicate business code %d found for business '%s' and '%s' in app '%s'",
					bizCode, bizName, existingBiz, appName)
			}
			seenCodes[bizCode] = bizName
		}
	}
	return nil
}

// getBizCode gets the business code for a given business name
func getBizCode(metadata map[string]interface{}, appName, bizName string) (int, error) {
	apps, ok := metadata["app"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid metadata format")
	}

	for _, appInterface := range apps {
		app, ok := appInterface.(map[string]interface{})
		if !ok {
			continue
		}

		if app["name"].(string) != appName {
			continue
		}

		businessList, ok := app["business"].([]interface{})
		if !ok {
			continue
		}

		for _, bizInterface := range businessList {
			biz, ok := bizInterface.(map[string]interface{})
			if !ok {
				continue
			}

			if biz["name"].(string) == bizName {
				return biz["code"].(int), nil
			}
		}
	}

	return 0, fmt.Errorf("business domain '%s' not found in app '%s'", bizName, appName)
}

// errorCode calculates the final error code
func errorCode(appCode, bizCode, subCode int) int {
	return appCode*100000000 + bizCode*10000 + subCode
}

// CodeGenData represents the data for code generation template
type CodeGenData struct {
	PackageName   string
	BizName       string
	AppName       string
	ImportPath    string
	Constants     []ConstantGroup
	Registrations []string
}

// ConstantGroup represents a group of constants for one error
type ConstantGroup struct {
	CodeName       string
	CodeValue      int
	MessageName    string
	MessageValue   string
	StabilityName  string
	StabilityValue string
	Description    string
}

// getString safely gets a string value from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getInt safely gets an int value from map
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

// getBool safely gets a bool value from map
func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// generateGoCode generates Go code for the given business domain
func generateGoCode(config *Config, bizName string, bizCode int, commonErrors, bizErrors []interface{}, outputDir string) (string, error) {
	if outputDir == "" {
		return "", fmt.Errorf("output_dir cannot be empty")
	}

	packageName := strings.ToLower(bizName)

	var constants []ConstantGroup
	var registrations []string

	// Process all errors (common + business specific)
	allErrors := append(commonErrors, bizErrors...)

	for _, errorInterface := range allErrors {
		errorMap, ok := errorInterface.(map[string]interface{})
		if !ok {
			continue
		}

		code := errorCode(config.AppCode, bizCode, getInt(errorMap, "code"))
		name := getString(errorMap, "name")
		if name == "" {
			continue
		}

		unexportName := strings.ToLower(name[:1]) + name[1:]
		message := getString(errorMap, "message")
		description := getString(errorMap, "description")
		noAffect := getBool(errorMap, "no_affect_stability")

		// Create constant group
		constantGroup := ConstantGroup{
			CodeName:       name + "Code",
			CodeValue:      code,
			MessageName:    unexportName + "Message",
			MessageValue:   message,
			StabilityName:  unexportName + "NoAffectStability",
			StabilityValue: strconv.FormatBool(noAffect),
			Description:    description,
		}
		constants = append(constants, constantGroup)

		// Create registration
		registration := fmt.Sprintf(`	code.Register(
		%sCode,
		%sMessage,
		code.WithAffectStability(!%sNoAffectStability),
	)`, name, unexportName, unexportName)
		registrations = append(registrations, registration)
	}

	// Prepare template data
	data := CodeGenData{
		PackageName:   filepath.Base(filepath.Dir(outputDir)),
		BizName:       bizName,
		AppName:       config.AppName,
		ImportPath:    config.ImportPath,
		Constants:     constants,
		Registrations: registrations,
	}

	// Generate Go file
	outputPath := filepath.Join(outputDir, packageName+".go")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	tmpl := template.Must(template.New("gocode").Funcs(funcMap).Parse(goCodeTemplate))
	if err := tmpl.Execute(file, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Run go fmt
	cmd := exec.Command("go", "fmt", outputPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: go fmt failed: %v\n", err)
	}

	return outputPath, nil
}

// generateBizCode generates code for a specific business domain
func generateBizCode(config *Config, bizName string, bizCode int, commonErrors []interface{}, outputDir string) (string, error) {
	// Get business specific errors if they exist
	var bizErrors []interface{}
	bizErrorFile := filepath.Join(config.ScriptDir, bizName+".yaml")
	if _, err := os.Stat(bizErrorFile); err == nil {
		bizData, err := loadYAML(bizErrorFile)
		if err != nil {
			return "", fmt.Errorf("failed to load business error file: %w", err)
		}
		if errorCodes, ok := bizData["error_code"].([]interface{}); ok {
			bizErrors = errorCodes
		}
	}

	// Determine output directory
	if outputDir == "" {
		// Use template to generate output path
		outputDir = strings.ReplaceAll(config.OutputTemplate, "{project_root}", config.ProjectRoot)
		outputDir = strings.ReplaceAll(outputDir, "{biz}", bizName)
	} else {
		outputDir = os.ExpandEnv(outputDir)
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(config.ProjectRoot, outputDir)
		}
	}

	return generateGoCode(config, bizName, bizCode, commonErrors, bizErrors, outputDir)
}

// goCodeTemplate is the template for generating Go code
const goCodeTemplate = `// Code generated by tool. DO NOT EDIT.
// app: {{.AppName}}, biz: {{.BizName}}

package {{.PackageName}}

import (
	"{{.ImportPath}}"
)

const (
{{- range $i, $const := .Constants}}
	{{$const.CodeName}} = {{$const.CodeValue}}{{if $const.Description}} // {{$const.Description}}{{end}}
	{{$const.MessageName}} = "{{$const.MessageValue}}"
	{{$const.StabilityName}} = {{$const.StabilityValue}}
{{if ne $i (len $.Constants | add -1)}}
{{end}}
{{- end}}
)

func init() {
{{range .Registrations}}
{{.}}

{{end}}
}
`

func main() {
	var (
		bizName        = flag.String("biz", "", "Business domain name (e.g., evaluation) or \"*\" to generate for all business domains")
		outputDir      = flag.String("output-dir", "", "Output directory for generated Go file")
		appName        = flag.String("app-name", "myapp", "Application name")
		appCode        = flag.Int("app-code", 1, "Application code (1-9)")
		importPath     = flag.String("import-path", "github.com/example/project/pkg/errorx/code", "Import path for error code package")
		scriptDir      = flag.String("script-dir", "", "Script directory (default: current directory)")
		projectRoot    = flag.String("project-root", "", "Project root directory (default: 3 levels up from script dir)")
		outputTemplate = flag.String("output-template", "{project_root}/modules/{biz}/pkg/errno", "Output directory template")
		metadataFile   = flag.String("metadata-file", "metadata.yaml", "Metadata file name")
		commonFile     = flag.String("common-file", "common.yaml", "Common errors file name")
	)
	flag.Parse()

	// Handle positional argument for biz name
	if *bizName == "" && len(flag.Args()) > 0 {
		*bizName = flag.Args()[0]
	}

	if *bizName == "" {
		fmt.Println("Usage: go run code_gen_simple.go [options] <biz>")
		fmt.Println("       go run code_gen_simple.go --biz <biz> [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  go run code_gen_simple.go evaluation")
		fmt.Println("  go run code_gen_simple.go --biz evaluation --app-name myapp --app-code 6")
		fmt.Println("  go run code_gen_simple.go \"*\" --output-template \"{project_root}/internal/{biz}/errors\"")
		os.Exit(1)
	}

	// Determine script directory
	if *scriptDir == "" {
		var err error
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
		*scriptDir = pwd
	}

	// Determine project root
	if *projectRoot == "" {
		*projectRoot = filepath.Dir(filepath.Dir(filepath.Dir(*scriptDir)))
	}

	// Validate app code
	if *appCode < 1 || *appCode > 9 {
		log.Fatal("App code must be between 1 and 9")
	}

	// Create config
	config := &Config{
		AppName:        *appName,
		AppCode:        *appCode,
		ImportPath:     *importPath,
		ScriptDir:      *scriptDir,
		ProjectRoot:    *projectRoot,
		OutputTemplate: *outputTemplate,
	}

	// Load configuration files
	metadataPath := filepath.Join(*scriptDir, *metadataFile)
	metadata, err := loadYAML(metadataPath)
	if err != nil {
		log.Fatalf("Failed to load metadata from %s: %v", metadataPath, err)
	}

	// Validate business codes
	if err := validateBusinessCodes(metadata); err != nil {
		log.Fatalf("Error in %s: %v", *metadataFile, err)
	}

	// Load common errors
	commonPath := filepath.Join(*scriptDir, *commonFile)
	commonData, err := loadYAML(commonPath)
	if err != nil {
		log.Fatalf("Failed to load common errors from %s: %v", commonPath, err)
	}

	commonErrors, ok := commonData["error_code"].([]interface{})
	if !ok {
		log.Fatal("Invalid common errors format: error_code field not found or not an array")
	}

	// Get target app from metadata
	apps, ok := metadata["app"].([]interface{})
	if !ok {
		log.Fatal("Invalid metadata format: app field not found")
	}

	var targetApp map[string]interface{}
	for _, appInterface := range apps {
		app, ok := appInterface.(map[string]interface{})
		if !ok {
			continue
		}
		if app["name"].(string) == *appName {
			targetApp = app
			break
		}
	}

	if targetApp == nil {
		log.Fatalf("Error: app '%s' not found in metadata", *appName)
	}

	// Handle wildcard case
	if *bizName == "*" {
		businessList, ok := targetApp["business"].([]interface{})
		if !ok {
			log.Fatal("Invalid app format: business field not found")
		}

		for _, bizInterface := range businessList {
			biz, ok := bizInterface.(map[string]interface{})
			if !ok {
				continue
			}

			bizName := biz["name"].(string)
			if bizName == "common" {
				continue
			}

			bizCode := biz["code"].(int)
			fmt.Printf("\nProcessing business domain: %s\n", bizName)
			result, err := generateBizCode(config, bizName, bizCode, commonErrors, *outputDir)
			if err != nil {
				log.Fatalf("Failed to generate code for %s: %v", bizName, err)
			}
			fmt.Printf("Generated error codes written to: %s\n", result)
		}
		return
	}

	// Handle single business domain case
	bizCode, err := getBizCode(metadata, *appName, *bizName)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	result, err := generateBizCode(config, *bizName, bizCode, commonErrors, *outputDir)
	if err != nil {
		log.Fatalf("Failed to generate code: %v", err)
	}
	fmt.Printf("Generated error codes written to: %s\n", result)
}
