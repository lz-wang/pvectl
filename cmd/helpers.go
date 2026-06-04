package cmd

import (
	"fmt"
	"strconv"

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
