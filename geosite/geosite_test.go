package geosite

import (
	"testing"

	"github.com/xtls/xray-core/app/router"
)

func TestAddGeoSiteEntryCreatesAndMerges(t *testing.T) {
	in := &router.GeoSiteList{}

	out, err := AddGeoSiteEntry(in, "my-sites", []string{"example.com", "example.com"}, router.Domain_Full)
	if err != nil {
		t.Fatalf("AddGeoSiteEntry create: %v", err)
	}
	if len(out.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entry))
	}
	if out.Entry[0].CountryCode != "MY-SITES" {
		t.Fatalf("expected normalized code, got %q", out.Entry[0].CountryCode)
	}
	if len(out.Entry[0].Domain) != 1 {
		t.Fatalf("expected duplicate domain to be ignored, got %d domains", len(out.Entry[0].Domain))
	}
	if out.Entry[0].Domain[0].Type != router.Domain_Full {
		t.Fatalf("expected domain type full, got %s", out.Entry[0].Domain[0].Type.String())
	}

	merged, err := AddGeoSiteEntry(out, "MY-SITES", []string{"example.org", "example.com"}, router.Domain_Domain)
	if err != nil {
		t.Fatalf("AddGeoSiteEntry merge: %v", err)
	}
	if len(merged.Entry) != 1 {
		t.Fatalf("expected merge into existing entry, got %d entries", len(merged.Entry))
	}
	if len(merged.Entry[0].Domain) != 3 {
		t.Fatalf("expected 3 unique domain rules after merge, got %d", len(merged.Entry[0].Domain))
	}
	if merged.Entry[0].Domain[1].Type != router.Domain_Domain {
		t.Fatalf("expected appended domain type domain, got %s", merged.Entry[0].Domain[1].Type.String())
	}
	if merged.Entry[0].Domain[2].Type != router.Domain_Domain {
		t.Fatalf("expected same value with different type to be kept, got %s", merged.Entry[0].Domain[2].Type.String())
	}
}
