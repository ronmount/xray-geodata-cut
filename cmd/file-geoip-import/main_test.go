package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCIDRs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ddos-guard.txt")
	content := strings.Join([]string{
		"# comment",
		"",
		"45.10.240.0/24",
		"203.0.113.5",
		"2001:db8::1",
		"45.10.240.0/24",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := loadCIDRs(path)
	if err != nil {
		t.Fatalf("loadCIDRs: %v", err)
	}

	want := []string{
		"45.10.240.0/24",
		"203.0.113.5/32",
		"2001:db8::1/128",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d CIDRs, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected CIDR %q at index %d, got %q", want[i], i, got[i])
		}
	}
}

func TestLoadCIDRsRejectsInvalidLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.txt")
	if err := os.WriteFile(path, []byte("not-an-ip\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := loadCIDRs(path)
	if err == nil {
		t.Fatal("expected loadCIDRs to reject invalid input")
	}
}

func TestDefaultCodeFromPath(t *testing.T) {
	if got := defaultCodeFromPath("./ddos-guard.txt"); got != "DDOS-GUARD" {
		t.Fatalf("expected DDOS-GUARD, got %q", got)
	}
}
