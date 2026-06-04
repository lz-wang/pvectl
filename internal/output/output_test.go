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

func TestWriteCloneResultJSON(t *testing.T) {
	var buf bytes.Buffer
	result := CloneResult{Kind: "vm", SourceVMID: 9000, NewVMID: 101, SourceNode: "pve1", TargetNode: "pve2", Name: "app-vm"}

	if err := WriteCloneResult(&buf, "json", result); err != nil {
		t.Fatalf("write clone json: %v", err)
	}
	if !strings.Contains(buf.String(), `"new_vmid": 101`) {
		t.Fatalf("json output = %s", buf.String())
	}
}

func TestWriteCloneResultTable(t *testing.T) {
	var buf bytes.Buffer
	result := CloneResult{Kind: "lxc", SourceVMID: 900, NewVMID: 201, SourceNode: "pve1", TargetNode: "pve1", Name: "app-lxc"}

	if err := WriteCloneResult(&buf, "table", result); err != nil {
		t.Fatalf("write clone table: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "NEWID") || !strings.Contains(got, "201") {
		t.Fatalf("table output = %s", got)
	}
}

func TestWriteSnapshotRowsJSON(t *testing.T) {
	var buf bytes.Buffer
	rows := []SnapshotRow{{Kind: "vm", VMID: 101, Node: "pve1", Name: "before-upgrade", Snaptime: 1710000000}}

	if err := WriteSnapshotRows(&buf, "json", rows); err != nil {
		t.Fatalf("write snapshot json: %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "before-upgrade"`) {
		t.Fatalf("json output = %s", buf.String())
	}
}

func TestWriteSnapshotRowsTable(t *testing.T) {
	var buf bytes.Buffer
	rows := []SnapshotRow{{Kind: "lxc", VMID: 201, Node: "pve1", Name: "before-upgrade", Parent: "base", Snaptime: 1710000000}}

	if err := WriteSnapshotRows(&buf, "table", rows); err != nil {
		t.Fatalf("write snapshot table: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "SNAPTIME") || !strings.Contains(got, "before-upgrade") {
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
