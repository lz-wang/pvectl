package cmd

import (
	"reflect"
	"testing"
)

func TestNormalizeArgsMovesLeafFlagsBeforePositionals(t *testing.T) {
	args := []string{"pvectl", "vm", "start", "100", "--wait", "--wait-timeout", "10s"}
	want := []string{"pvectl", "vm", "start", "--wait", "--wait-timeout", "10s", "100"}

	if got := normalizeArgs(args); !reflect.DeepEqual(got, want) {
		t.Fatalf("normalize = %#v, want %#v", got, want)
	}
}

func TestNormalizeArgsMovesOutputFlag(t *testing.T) {
	args := []string{"pvectl", "vm", "get", "100", "-o", "json"}
	want := []string{"pvectl", "vm", "get", "-o", "json", "100"}

	if got := normalizeArgs(args); !reflect.DeepEqual(got, want) {
		t.Fatalf("normalize = %#v, want %#v", got, want)
	}
}

func TestNormalizeArgsMovesRepeatedSetFlags(t *testing.T) {
	args := []string{"pvectl", "vm", "config", "101", "--set", "memory=4096", "--set", "cores=4", "--wait"}
	want := []string{"pvectl", "vm", "config", "--set", "memory=4096", "--set", "cores=4", "--wait", "101"}

	if got := normalizeArgs(args); !reflect.DeepEqual(got, want) {
		t.Fatalf("normalize = %#v, want %#v", got, want)
	}
}
