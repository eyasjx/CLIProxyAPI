package thinking

import (
	"strings"

	"github.com/MoeForge/moego/ojson"
)

// OpenAIThinkingTypeForAntigravity reads the optional thinking.type alias while
// leaving explicit reasoning_effort and requests without an alias authoritative.
func OpenAIThinkingTypeForAntigravity(body []byte) string {
	if ojson.GetBytesNoCopy(body, "reasoning_effort").Exists() {
		return ""
	}
	for _, path := range []string{"thinking.type", "extra_body.thinking.type"} {
		value := ojson.GetBytesNoCopy(body, path)
		if value.Type != ojson.String {
			continue
		}
		switch normalized := strings.ToLower(strings.TrimSpace(value.String())); normalized {
		case "enabled", "disabled":
			return normalized
		default:
			return ""
		}
	}
	return ""
}
