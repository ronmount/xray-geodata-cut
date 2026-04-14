package redko

import (
	"strings"
	"testing"
)

func TestParseIPBlocks(t *testing.T) {
	input := `
2gis
91.236.51.50
91.221.199.120
yandex
91.236.51.145
2001:db8::1
2gis
91.236.49.6
`

	blocks, err := ParseIPBlocks(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseIPBlocks: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Key != "2GIS" {
		t.Fatalf("expected first key 2GIS, got %q", blocks[0].Key)
	}
	if len(blocks[0].CIDRs) != 3 {
		t.Fatalf("expected 3 CIDRs in first block, got %d", len(blocks[0].CIDRs))
	}
	if blocks[0].CIDRs[0] != "91.236.51.50/32" {
		t.Fatalf("expected /32 conversion, got %q", blocks[0].CIDRs[0])
	}
	if blocks[1].CIDRs[1] != "2001:db8::1/128" {
		t.Fatalf("expected /128 conversion, got %q", blocks[1].CIDRs[1])
	}
}

func TestParseIPBlocksRejectsIPBeforeKey(t *testing.T) {
	_, err := ParseIPBlocks(strings.NewReader("91.236.51.50\n"))
	if err == nil {
		t.Fatal("expected error for IP before key")
	}
}
