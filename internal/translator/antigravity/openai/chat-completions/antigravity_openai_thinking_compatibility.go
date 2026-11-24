package chat_completions

import (
	"strings"

	"github.com/MoeForge/moego/ojson"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/thinking"
)

// applyOpenAIThinkingCompatibilityToAntigravityDoc supplements upstream behavior
// only when the client supplies a thinking.type alias without reasoning_effort.
func applyOpenAIThinkingCompatibilityToAntigravityDoc(outDoc *ojson.Doc, rawJSON []byte, modelName string) string {
	compatibilityType := thinking.OpenAIThinkingTypeForAntigravity(rawJSON)
	switch compatibilityType {
	case "enabled":
		_ = outDoc.Set("request.generationConfig.thinkingConfig.thinkingLevel", string(thinking.LevelHigh))
		_ = outDoc.Set("request.generationConfig.thinkingConfig.includeThoughts", true)
	case "disabled":
		_ = outDoc.Set("request.generationConfig.thinkingConfig.thinkingLevel", antigravityOpenAILowestThinkingLevel(modelName))
		_ = outDoc.Set("request.generationConfig.thinkingConfig.includeThoughts", false)
	default:
		return ""
	}
	return compatibilityType
}

func antigravityOpenAILowestThinkingLevel(modelName string) string {
	baseModel := thinking.ParseSuffix(modelName).ModelName
	if modelInfo := registry.LookupModelInfo(baseModel, "antigravity"); modelInfo != nil && modelInfo.Thinking != nil {
		for _, level := range modelInfo.Thinking.Levels {
			if normalized := strings.ToLower(strings.TrimSpace(level)); normalized != "" {
				return normalized
			}
		}
	}
	return string(thinking.LevelMinimal)
}
