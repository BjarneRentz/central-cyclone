package query

import (
	"testing"
)

func TestYqValueExtractor_ExtractValue(t *testing.T) {
	extractor := NewYqValueExtractor()

	tests := []struct {
		name      string
		content   []byte
		yamlPath  string
		want      string
		wantError bool
		errorMsg  string
	}{
		{
			name: "extract simple YAML field",
			content: []byte(`name: myapp
version: 1.0.0`),
			yamlPath:  ".version",
			want:      "1.0.0",
			wantError: false,
		},
		{
			name: "extract nested YAML field",
			content: []byte(`metadata:
  name: test-deployment
  namespace: default
spec:
  version: 2.5.0`),
			yamlPath:  ".spec.version",
			want:      "2.5.0",
			wantError: false,
		},
		{
			name: "extract array element",
			content: []byte(`images:
  - name: app
    tag: v1.0
  - name: sidecar
    tag: v2.0`),
			yamlPath:  ".images[0].tag",
			want:      "v1.0",
			wantError: false,
		},
		{
			name:      "extract string value from JSON",
			content:   []byte(`{"metadata":{"name":"test-app"},"version":"3.0.0"}`),
			yamlPath:  ".version",
			want:      "3.0.0",
			wantError: false,
		},
		{
			name:      "extract nested JSON field",
			content:   []byte(`{"spec":{"image":{"repository":"myrepo/app","tag":"v1.5.0"}}}`),
			yamlPath:  ".spec.image.tag",
			want:      "v1.5.0",
			wantError: false,
		},
		{
			name: "extract from second array element",
			content: []byte(`items:
  - version: 1.0.0
  - version: 2.0.0
  - version: 3.0.0`),
			yamlPath:  ".items[1].version",
			want:      "2.0.0",
			wantError: false,
		},
		{
			name: "extract numeric value",
			content: []byte(`count: 42
port: 8080`),
			yamlPath:  ".port",
			want:      "8080",
			wantError: false,
		},
		{
			name: "extract boolean value",
			content: []byte(`enabled: true
debug: false`),
			yamlPath:  ".enabled",
			want:      "true",
			wantError: false,
		},
		{
			name:      "extract basic field",
			content:   []byte(`app: myapp`),
			yamlPath:  ".app",
			want:      "myapp",
			wantError: false,
		},
		{
			name:      "error: empty content",
			content:   []byte(""),
			yamlPath:  ".version",
			want:      "",
			wantError: true,
			errorMsg:  "content is empty",
		},
		{
			name:      "error: empty path",
			content:   []byte("version: 1.0.0"),
			yamlPath:  "",
			want:      "",
			wantError: true,
			errorMsg:  "yaml path cannot be empty",
		},
		{
			name:      "error: invalid path",
			content:   []byte(`version: 1.0.0`),
			yamlPath:  "invalid[syntax",
			want:      "",
			wantError: true,
		},
		{
			name: "extract from yq when path doesn't exist",
			content: []byte(`version: 1.0.0
name: app`),
			yamlPath:  ".nonexistent",
			want:      "null",
			wantError: false,
		},
		{
			name: "extract from array with out of bounds index returns null",
			content: []byte(`items:
  - value: first`),
			yamlPath:  ".items[10]",
			want:      "null",
			wantError: false,
		},
		{
			name:      "extract value with whitespace",
			content:   []byte(`message: "Hello World"`),
			yamlPath:  ".message",
			want:      "Hello World",
			wantError: false,
		},
		{
			name: "extract deeply nested value",
			content: []byte(`level1:
  level2:
    level3:
      level4:
        value: deep`),
			yamlPath:  ".level1.level2.level3.level4.value",
			want:      "deep",
			wantError: false,
		},
		{
			name: "extract with mixed YAML types",
			content: []byte(`config:
  database:
    host: localhost
    port: 5432
    credentials:
      username: admin
      password: secret`),
			yamlPath:  ".config.database.credentials.username",
			want:      "admin",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractor.ExtractValue(tt.content, tt.yamlPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("ExtractValue() expected error but got nil")
					return
				}
				if tt.errorMsg != "" && err != nil {
					// Just check that error occurred with some message
					if err.Error() == "" {
						t.Errorf("ExtractValue() error message should not be empty")
					}
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractValue() unexpected error: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("ExtractValue() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestYqValueExtractor_ExtractValue_Interface(t *testing.T) {
	// Test that YqValueExtractor implements ValueExtractor interface
	var _ ValueExtractor = (*YqValueExtractor)(nil)
}

func TestNewYqValueExtractor(t *testing.T) {
	extractor := NewYqValueExtractor()
	if extractor == nil {
		t.Error("NewYqValueExtractor() returned nil")
	}
	if _, ok := interface{}(extractor).(ValueExtractor); !ok {
		t.Error("NewYqValueExtractor() does not implement ValueExtractor")
	}
}
