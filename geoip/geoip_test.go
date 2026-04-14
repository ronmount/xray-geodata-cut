package geoip

import (
	"testing"

	"github.com/xtls/xray-core/app/router"
)

func TestAddGeoIPEntryCreatesAndMerges(t *testing.T) {
	in := &router.GeoIPList{}

	out, err := AddGeoIPEntry(in, "ru-custom", []string{"203.0.113.0/24", "203.0.113.0/24"}, true)
	if err != nil {
		t.Fatalf("AddGeoIPEntry create: %v", err)
	}
	if len(out.Entry) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entry))
	}
	if out.Entry[0].CountryCode != "RU-CUSTOM" {
		t.Fatalf("expected normalized code, got %q", out.Entry[0].CountryCode)
	}
	if !out.Entry[0].ReverseMatch {
		t.Fatalf("expected reverse match to be set on new entry")
	}
	if len(out.Entry[0].Cidr) != 1 {
		t.Fatalf("expected duplicate CIDR to be ignored, got %d CIDRs", len(out.Entry[0].Cidr))
	}

	merged, err := AddGeoIPEntry(out, "RU-CUSTOM", []string{"198.51.100.0/24", "203.0.113.0/24"}, false)
	if err != nil {
		t.Fatalf("AddGeoIPEntry merge: %v", err)
	}
	if len(merged.Entry) != 1 {
		t.Fatalf("expected merge into existing entry, got %d entries", len(merged.Entry))
	}
	if len(merged.Entry[0].Cidr) != 2 {
		t.Fatalf("expected 2 unique CIDRs after merge, got %d", len(merged.Entry[0].Cidr))
	}
	if !merged.Entry[0].ReverseMatch {
		t.Fatalf("expected existing reverse match to stay unchanged")
	}
}
