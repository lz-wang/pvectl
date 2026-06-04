package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

func parseVMID(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("vmid is required")
	}
	vmid, err := strconv.Atoi(value)
	if err != nil || vmid <= 0 {
		return 0, fmt.Errorf("invalid vmid %q", value)
	}
	return vmid, nil
}

func requireNoExtraArgs(c *cli.Context, want int) error {
	if c.NArg() == want {
		return nil
	}
	return fmt.Errorf("expected %d argument(s), got %d", want, c.NArg())
}

func parseSetFlags(items []string) (map[string]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one --set key=value is required")
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --set %q, expected key=value", item)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in --set %q", item)
		}
		result[key] = value
	}
	return result, nil
}
