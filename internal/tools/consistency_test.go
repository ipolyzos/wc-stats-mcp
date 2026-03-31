package tools

import (
	"encoding/json"
	"slices"
	"testing"

	"wc-stats-mcp/internal/validate"
)

func TestRegisterProduction_matchesProductionToolNames(t *testing.T) {
	r := NewRegistry()
	RegisterProduction(r, nil, validate.Limits{})
	got := r.Names()
	if !slices.Equal(got, ProductionToolNames) {
		t.Fatalf("registry names %v, want %v", got, ProductionToolNames)
	}
}

func TestToolMeta_schemasMarshalAndUnmarshal(t *testing.T) {
	for _, name := range ProductionToolNames {
		_, schema := ToolMeta(name)
		b, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("%s: marshal schema: %v", name, err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatalf("%s: unmarshal schema: %v", name, err)
		}
		b2, err := json.Marshal(decoded)
		if err != nil {
			t.Fatalf("%s: remarshal: %v", name, err)
		}
		var round map[string]any
		if err := json.Unmarshal(b2, &round); err != nil {
			t.Fatalf("%s: second unmarshal: %v", name, err)
		}
		ty, ok := round["type"].(string)
		if !ok || ty != "object" {
			t.Fatalf("%s: schema.type = %q, want object", name, ty)
		}
	}
}

func TestToolMeta_unknownNameStillValidJSONSchemaShape(t *testing.T) {
	_, schema := ToolMeta("tool_not_registered")
	b, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
}
