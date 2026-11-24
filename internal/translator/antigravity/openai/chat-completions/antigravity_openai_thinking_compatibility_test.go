package chat_completions

import (
	"testing"

	"github.com/MoeForge/moego/ojson"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
)

func TestConvertOpenAIRequestToAntigravityMapsThinkingType(t *testing.T) {
	tests := []struct {
		name             string
		body             string
		wantLevel        string
		wantInclude      bool
		wantIncludeExist bool
		wantConfigAbsent bool
	}{
		{
			name:             "enabled defaults to high and includes thoughts",
			body:             `{"model":"gemini-3.6-flash-high","thinking":{"type":"enabled"}}`,
			wantLevel:        "high",
			wantInclude:      true,
			wantIncludeExist: true,
		},
		{
			name:             "disabled uses the lowest supported level and hides thoughts",
			body:             `{"model":"gemini-3.6-flash-high","thinking":{"type":"disabled"}}`,
			wantLevel:        "minimal",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "nested extra body thinking type is supported",
			body:             `{"model":"gemini-3.6-flash-high","extra_body":{"thinking":{"type":"disabled"}}}`,
			wantLevel:        "minimal",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "explicit reasoning effort wins over enabled type",
			body:             `{"model":"gemini-3.6-flash-high","reasoning_effort":"low","thinking":{"type":"enabled"}}`,
			wantLevel:        "low",
			wantInclude:      true,
			wantIncludeExist: true,
		},
		{
			name:             "explicit reasoning effort wins over disabled type",
			body:             `{"model":"gemini-3.6-flash-high","reasoning_effort":"low","thinking":{"type":"disabled"}}`,
			wantLevel:        "low",
			wantInclude:      true,
			wantIncludeExist: true,
		},
		{
			name:             "omitted controls preserve upstream defaults",
			body:             `{"model":"gemini-3.6-flash-high"}`,
			wantConfigAbsent: true,
		},
		{
			name:             "omitted type preserves explicit generation config",
			body:             `{"generationConfig":{"thinkingConfig":{"thinkingLevel":"low","includeThoughts":false}}}`,
			wantLevel:        "low",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "reasoning effort alone preserves upstream summaries",
			body:             `{"reasoning_effort":"high"}`,
			wantLevel:        "high",
			wantInclude:      true,
			wantIncludeExist: true,
		},
		{
			name:             "disabled cannot be overridden by explicit summary visibility",
			body:             `{"thinking":{"type":"disabled","includeThoughts":true}}`,
			wantLevel:        "minimal",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "enabled preserves explicit summary visibility",
			body:             `{"thinking":{"type":"enabled","includeThoughts":false}}`,
			wantLevel:        "high",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "auto reasoning effort wins over disabled type",
			body:             `{"reasoning_effort":"auto","thinking":{"type":"disabled"}}`,
			wantInclude:      true,
			wantIncludeExist: true,
		},
		{
			name:             "none reasoning effort wins over enabled type",
			body:             `{"reasoning_effort":"none","thinking":{"type":"enabled"}}`,
			wantLevel:        "none",
			wantInclude:      false,
			wantIncludeExist: true,
		},
		{
			name:             "empty explicit reasoning effort preserves upstream behavior",
			body:             `{"reasoning_effort":"","thinking":{"type":"disabled"}}`,
			wantConfigAbsent: true,
		},
		{
			name:             "unrecognized type preserves upstream defaults",
			body:             `{"thinking":{"type":"unknown"}}`,
			wantConfigAbsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, errConvert := ConvertOpenAIRequestToAntigravity("gemini-3.6-flash-high", []byte(tt.body), false)
			if errConvert != nil {
				t.Fatalf("ConvertOpenAIRequestToAntigravity() error = %v", errConvert)
			}
			level := ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig.thinkingLevel")
			if got := level.String(); got != tt.wantLevel {
				t.Fatalf("thinkingLevel = %q, want %q. Output: %s", got, tt.wantLevel, out)
			}
			if tt.wantLevel == "" && level.Exists() {
				t.Fatalf("thinkingLevel should be absent. Output: %s", out)
			}
			includeThoughts := ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig.includeThoughts")
			if includeThoughts.Exists() != tt.wantIncludeExist {
				t.Fatalf("includeThoughts exists = %v, want %v. Output: %s", includeThoughts.Exists(), tt.wantIncludeExist, out)
			}
			if includeThoughts.Bool() != tt.wantInclude {
				t.Fatalf("includeThoughts = %v, want %v. Output: %s", includeThoughts.Bool(), tt.wantInclude, out)
			}
			if tt.wantConfigAbsent && ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig").Exists() {
				t.Fatalf("thinkingConfig should be absent without controls. Output: %s", out)
			}
		})
	}
}

func TestConvertOpenAIRequestToAntigravityDisabledUsesModelCapabilities(t *testing.T) {
	const modelID = "gemini-thinking-compatibility-low"
	const clientID = "thinking-compatibility-low-client"
	reg := registry.GetGlobalRegistry()
	reg.RegisterClient(clientID, "antigravity", []*registry.ModelInfo{{
		ID:       modelID,
		Thinking: &registry.ThinkingSupport{Levels: []string{"low", "high"}},
	}})
	t.Cleanup(func() { reg.UnregisterClient(clientID) })

	out, errConvert := ConvertOpenAIRequestToAntigravity(modelID+"(high)", []byte(`{"thinking":{"type":"disabled"}}`), false)
	if errConvert != nil {
		t.Fatalf("ConvertOpenAIRequestToAntigravity() error = %v", errConvert)
	}
	if got := ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig.thinkingLevel").String(); got != "low" {
		t.Fatalf("thinkingLevel = %q, want low. Output: %s", got, out)
	}
	if got := ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig.includeThoughts"); got.Type != ojson.False {
		t.Fatalf("includeThoughts = %s, want false. Output: %s", got.Raw, out)
	}
}
