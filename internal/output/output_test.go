package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateFormat(t *testing.T) {
	for _, format := range []string{"table", "json", "yaml"} {
		if err := ValidateFormat(format); err != nil {
			t.Fatalf("format %q: %v", format, err)
		}
	}
	if err := ValidateFormat("xml"); err == nil {
		t.Fatal("expected invalid format")
	}
}

func TestWriteGuestRowsJSON(t *testing.T) {
	var buf bytes.Buffer
	rows := []GuestRow{{Kind: "vm", VMID: 100, Name: "debian", Node: "pve1", Status: "running"}}

	if err := WriteGuestRows(&buf, "json", rows); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if !strings.Contains(buf.String(), `"vmid": 100`) {
		t.Fatalf("json output = %s", buf.String())
	}
}

func TestWriteGuestRowsTable(t *testing.T) {
	var buf bytes.Buffer
	rows := []GuestRow{{Kind: "vm", VMID: 100, Name: "debian", Node: "pve1", Status: "running", MaxMem: 1024 * 1024 * 1024}}

	if err := WriteGuestRows(&buf, "table", rows); err != nil {
		t.Fatalf("write table: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "VMID") || !strings.Contains(got, "1.0GiB") {
		t.Fatalf("table output = %s", got)
	}
}

func TestFormatUptime(t *testing.T) {
	if got := FormatUptime(0); got != "-" {
		t.Fatalf("zero uptime = %q", got)
	}
	if got := FormatUptime(3*24*60*60 + 4*60*60 + 5*60); got != "3d4h5m" {
		t.Fatalf("uptime = %q", got)
	}
}
