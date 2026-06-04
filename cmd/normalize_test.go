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
