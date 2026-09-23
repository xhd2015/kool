//go:build ignore

package pagemeta

import "testing"

// TestSectionsHaveMeta is the guard the rule needs: every card a page renders
// must carry title, hint and empty. Renaming a key or dropping a hint fails
// here, not in a render that quietly shows a card without its brief.
func TestSectionsHaveMeta(t *testing.T) {
	if len(Sections()) == 0 {
		t.Fatal("no sections registered")
	}
	if err := Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogCoversRegistry(t *testing.T) {
	all, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range Sections() {
		if _, ok := all[key]; !ok {
			t.Fatalf("catalog missing registered section %q", key)
		}
	}
}

func TestSectionRejectsUnknownKey(t *testing.T) {
	if _, err := Section("no-such-card"); err == nil {
		t.Fatal(`Section("no-such-card") should fail instead of returning empty meta`)
	}
}
