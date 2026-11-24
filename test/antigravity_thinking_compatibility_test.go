package test

import (
	"testing"

	"github.com/MoeForge/moego/ojson"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	_ "github.com/router-for-me/CLIProxyAPI/v8/internal/thinking/provider/antigravity"
	_ "github.com/router-for-me/CLIProxyAPI/v8/internal/translator"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
)

func TestAntigravityThinkingCompatibilityFinalPipeline(t *testing.T) {
	const model = "gemini-3.6-flash-high"
	modelInfo := &registry.ModelInfo{
		ID:       model,
		Thinking: &registry.ThinkingSupport{Levels: []string{"minimal", "low", "high"}},
	}
	tests := []struct {
		name        string
		body        string
		wantLevel   string
		wantInclude string
	}{
		{name: "disabled overrides visible source summary", body: `{"thinking":{"type":"disabled","includeThoughts":true}}`, wantLevel: "minimal", wantInclude: "false"},
		{name: "nested disabled overrides visible source summary", body: `{"extra_body":{"thinking":{"type":"disabled"}},"reasoning":{"exclude":false}}`, wantLevel: "minimal", wantInclude: "false"},
		{name: "explicit effort wins", body: `{"thinking":{"type":"disabled"},"reasoning_effort":"high"}`, wantLevel: "high", wantInclude: "true"},
		{name: "enabled preserves hidden source summary", body: `{"thinking":{"type":"enabled","includeThoughts":false}}`, wantLevel: "high", wantInclude: "false"},
		{name: "omitted controls preserve defaults", body: `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.body)
			translated := sdktranslator.TranslateRequest(sdktranslator.FormatOpenAI, sdktranslator.FormatAntigravity, model, source, false)
			out, errThinking := helps.ApplyRequestThinkingWithModelInfo(translated,
				cliproxyexecutor.Request{Model: model, Payload: source},
				cliproxyexecutor.Options{OriginalRequest: source},
				"openai", "antigravity", "antigravity", modelInfo)
			if errThinking != nil {
				t.Fatalf("ApplyRequestThinkingWithModelInfo() error = %v", errThinking)
			}
			config := ojson.GetBytesNoCopy(out, "request.generationConfig.thinkingConfig")
			if tt.wantLevel == "" {
				if config.Exists() {
					t.Fatalf("thinkingConfig should be absent without controls. Output: %s", out)
				}
				return
			}
			if got := config.Get("thinkingLevel").String(); got != tt.wantLevel {
				t.Fatalf("thinkingLevel = %q, want %q. Output: %s", got, tt.wantLevel, out)
			}
			if got := config.Get("includeThoughts").Raw; got != tt.wantInclude {
				t.Fatalf("includeThoughts = %q, want %q. Output: %s", got, tt.wantInclude, out)
			}
		})
	}
}
