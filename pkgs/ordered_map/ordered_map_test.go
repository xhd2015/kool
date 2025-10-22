package ordered_map

import (
	"encoding/json"
	"testing"
)

func TestOrderedMap(t *testing.T) {
	om := NewOrderedMap()
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)

	data, err := json.Marshal(om)
	if err != nil {
		t.Fatalf("failed to marshal ordered map: %v", err)
	}
	expectData := `{"a":1,"b":2,"c":3}`
	if string(data) != expectData {
		t.Fatalf("expected %s, got %s", expectData, string(data))
	}

	om2 := NewOrderedMap()
	om2.Set("a", 1)
	om2.Set("c", 3)
	om2.Set("b", 2)
	data2, err := json.Marshal(om2)
	if err != nil {
		t.Fatalf("failed to marshal ordered map: %v", err)
	}
	expectData2 := `{"a":1,"c":3,"b":2}`
	if string(data2) != expectData2 {
		t.Fatalf("expected %s, got %s", expectData2, string(data2))
	}
}
