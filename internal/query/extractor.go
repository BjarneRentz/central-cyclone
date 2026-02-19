package query

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type ValueExtractor interface {
	ExtractValue(content []byte, yamlPath string) (string, error)
}

type YqValueExtractor struct{}

func NewYqValueExtractor() *YqValueExtractor {
	return &YqValueExtractor{}
}

// ExtractValue extracts a value from YAML/JSON content using a yq expression
// Examples:
//   - ".metadata.name" -> gets the name field
//   - ".spec.version" -> gets the version
//   - ".items[0].image" -> gets the image from first array element
func (e *YqValueExtractor) ExtractValue(content []byte, yamlPath string) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("content is empty")
	}
	if yamlPath == "" {
		return "", fmt.Errorf("yaml path cannot be empty")
	}

	// Create a decoder for the input
	decoder := yqlib.NewYamlDecoder(yqlib.ConfiguredYamlPreferences)

	// Read the input documents
	docs, err := yqlib.ReadDocuments(bytes.NewReader(content), decoder)
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	// Ensure we have at least one document
	if docs.Len() == 0 {
		return "", fmt.Errorf("no documents found in content")
	}

	// Create a BufferWriter to capture output
	var output bytes.Buffer

	// Create an encoder for output
	encoder := yqlib.NewYamlEncoder(yqlib.ConfiguredYamlPreferences)

	// Create a printer writer and printer to capture results
	printerWriter := yqlib.NewSinglePrinterWriter(&output)
	printer := yqlib.NewPrinter(encoder, printerWriter)

	// Create and initialize the expression parser
	yqlib.InitExpressionParser()

	// Parse the expression
	node, err := yqlib.ExpressionParser.ParseExpression(yamlPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse yaml path '%s': %w", yamlPath, err)
	}

	// Create evaluator
	evaluator := yqlib.NewStreamEvaluator()

	// Evaluate the expression
	_, err = evaluator.Evaluate("", bytes.NewReader(content), node, printer, decoder)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate yaml path '%s': %w", yamlPath, err)
	}

	result := strings.TrimSpace(output.String())
	if result == "" {
		return "", fmt.Errorf("yaml path '%s' returned no results", yamlPath)
	}

	return result, nil
}
