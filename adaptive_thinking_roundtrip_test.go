package anthropic_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

// TestAdaptiveThinkingRoundTrip verifies that adaptive thinking config
// survives an unmarshal->marshal round-trip through the SDK.
// Before the fix, the "adaptive" discriminator was missing from the
// ThinkingConfigParamUnion registration, causing UnmarshalJSON to silently
// fail and omitzero to strip the zero-valued Thinking field on re-marshal.
func TestAdaptiveThinkingRoundTrip(t *testing.T) {
	input := `{
		"model": "claude-opus-4-6-20250415",
		"max_tokens": 16000,
		"thinking": {"type": "adaptive"},
		"output_config": {"effort": "high"},
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	var params anthropic.MessageNewParams
	if err := params.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if params.Thinking.OfAdaptive == nil {
		t.Fatal("Thinking.OfAdaptive is nil after unmarshal -- adaptive discriminator not registered")
	}

	if params.OutputConfig.Effort != "high" {
		t.Fatalf("OutputConfig.Effort = %q, want %q", params.OutputConfig.Effort, "high")
	}

	out, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	outStr := string(out)

	if !strings.Contains(outStr, `"thinking"`) {
		t.Fatalf("re-marshaled JSON is missing \"thinking\" field:\n%s", outStr)
	}
	if !strings.Contains(outStr, `"adaptive"`) {
		t.Fatalf("re-marshaled JSON is missing \"adaptive\" type:\n%s", outStr)
	}
	if !strings.Contains(outStr, `"output_config"`) {
		t.Fatalf("re-marshaled JSON is missing \"output_config\" field:\n%s", outStr)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	thinking, ok := result["thinking"].(map[string]any)
	if !ok {
		t.Fatalf("thinking is not an object in output: %v", result["thinking"])
	}
	if thinking["type"] != "adaptive" {
		t.Fatalf("thinking.type = %v, want \"adaptive\"", thinking["type"])
	}

	outputConfig, ok := result["output_config"].(map[string]any)
	if !ok {
		t.Fatalf("output_config is not an object in output: %v", result["output_config"])
	}
	if outputConfig["effort"] != "high" {
		t.Fatalf("output_config.effort = %v, want \"high\"", outputConfig["effort"])
	}

	t.Logf("Round-trip successful. Output JSON:\n%s", outStr)
}

// TestBudgetThinkingStillWorks ensures budget-based thinking (type: "enabled")
// is not regressed by the adaptive discriminator addition.
func TestBudgetThinkingStillWorks(t *testing.T) {
	input := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 16000,
		"thinking": {"type": "enabled", "budget_tokens": 10000},
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	var params anthropic.MessageNewParams
	if err := params.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if params.Thinking.OfEnabled == nil {
		t.Fatal("Thinking.OfEnabled is nil after unmarshal")
	}
	if params.Thinking.OfEnabled.BudgetTokens != 10000 {
		t.Fatalf("BudgetTokens = %d, want 10000", params.Thinking.OfEnabled.BudgetTokens)
	}

	out, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	outStr := string(out)
	if !strings.Contains(outStr, `"enabled"`) {
		t.Fatalf("re-marshaled JSON missing \"enabled\":\n%s", outStr)
	}
	if !strings.Contains(outStr, `"budget_tokens"`) {
		t.Fatalf("re-marshaled JSON missing \"budget_tokens\":\n%s", outStr)
	}

	t.Logf("Budget thinking round-trip successful. Output JSON:\n%s", outStr)
}

// TestDisabledThinkingStillWorks ensures disabled thinking is not regressed.
func TestDisabledThinkingStillWorks(t *testing.T) {
	input := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 16000,
		"thinking": {"type": "disabled"},
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	var params anthropic.MessageNewParams
	if err := params.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if params.Thinking.OfDisabled == nil {
		t.Fatal("Thinking.OfDisabled is nil after unmarshal")
	}

	out, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	outStr := string(out)
	if !strings.Contains(outStr, `"disabled"`) {
		t.Fatalf("re-marshaled JSON missing \"disabled\":\n%s", outStr)
	}

	t.Logf("Disabled thinking round-trip successful. Output JSON:\n%s", outStr)
}
