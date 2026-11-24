package thinking

import "testing"

func TestExtractTranslatedSummaryConfigAntigravityThinkingAlias(t *testing.T) {
	tests := []struct {
		name   string
		target string
		body   string
		want   SummaryMode
	}{
		{name: "disabled overrides visibility for antigravity", target: "antigravity", body: `{"thinking":{"type":"disabled","includeThoughts":true}}`, want: SummaryDisabled},
		{name: "nested disabled is normalized", target: "antigravity", body: `{"extra_body":{"thinking":{"type":" DISABLED "}}}`, want: SummaryDisabled},
		{name: "explicit effort wins", target: "antigravity", body: `{"reasoning_effort":"high","thinking":{"type":"disabled"}}`, want: SummaryEnabled},
		{name: "empty explicit effort wins", target: "antigravity", body: `{"reasoning_effort":"","thinking":{"type":"disabled"}}`, want: SummaryUnspecified},
		{name: "enabled without explicit visibility leaves summary unspecified", target: "antigravity", body: `{"thinking":{"type":"enabled"}}`, want: SummaryUnspecified},
		{name: "top level alias takes precedence", target: "antigravity", body: `{"thinking":{"type":"enabled","includeThoughts":true},"extra_body":{"thinking":{"type":"disabled"}}}`, want: SummaryEnabled},
		{name: "unknown top level alias does not fall through", target: "antigravity", body: `{"thinking":{"type":"unknown","includeThoughts":true},"extra_body":{"thinking":{"type":"disabled"}}}`, want: SummaryEnabled},
		{name: "gemini keeps official explicit visibility", target: "gemini", body: `{"thinking":{"type":"disabled","includeThoughts":true}}`, want: SummaryEnabled},
		{name: "claude keeps official explicit visibility", target: "claude", body: `{"thinking":{"type":"disabled","includeThoughts":true}}`, want: SummaryEnabled},
		{name: "omitted controls stay unspecified", target: "antigravity", body: `{}`, want: SummaryUnspecified},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractTranslatedSummaryConfig([]byte(tt.body), "openai", tt.target); got.Mode != tt.want {
				t.Fatalf("summary mode = %v, want %v", got.Mode, tt.want)
			}
		})
	}
}
