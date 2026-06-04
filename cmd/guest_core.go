package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/lz-wang/pvectl/internal/output"
)

func guestListCommand(kind string, deps Dependencies) *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "List guests",
		Flags: append(commonNodeFlag(), commonOutputFlags()...),
		Action: func(c *cli.Context) error {
			if err := requireNoExtraArgs(c, 0); err != nil {
				return err
			}
			rt, err := buildRuntime(c, deps)
			if err != nil {
				return err
			}
			svc := guestService(kind, rt)
			rows, err := svc.List(c.Context, c.String("node"))
			if err != nil {
				return err
			}
			return output.WriteGuestRows(rt.stdout, rt.format, rows)
		},
	}
}

func guestGetCommand(kind string, deps Dependencies) *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Show guest details",
		ArgsUsage: "VMID",
		Flags:     append(commonNodeFlag(), commonOutputFlags()...),
		Action: func(c *cli.Context) error {
			if err := requireNoExtraArgs(c, 1); err != nil {
				return err
			}
			vmid, err := parseVMID(c.Args().First())
			if err != nil {
				return err
			}
			rt, err := buildRuntime(c, deps)
			if err != nil {
				return err
			}
			svc := guestService(kind, rt)
			row, err := svc.Get(c.Context, vmid, c.String("node"))
			if err != nil {
				return err
			}
			return output.WriteGuestDetail(rt.stdout, rt.format, row)
		},
	}
}

func guestControlCommand(kind, name, usage string, deps Dependencies) *cli.Command {
	return &cli.Command{
		Name:      name,
		Usage:     usage,
		ArgsUsage: "VMID",
		Flags:     commonControlFlags(),
		Action: func(c *cli.Context) error {
			if err := requireNoExtraArgs(c, 1); err != nil {
				return err
			}
			vmid, err := parseVMID(c.Args().First())
			if err != nil {
				return err
			}
			rt, err := buildRuntime(c, deps)
			if err != nil {
				return err
			}
			svc := guestService(kind, rt)
			switch name {
			case "start":
				return svc.Start(c.Context, vmid, c.String("node"))
			case "shutdown":
				return svc.Shutdown(c.Context, vmid, c.String("node"))
			case "stop":
				return svc.Stop(c.Context, vmid, c.String("node"))
			case "reboot":
				return svc.Reboot(c.Context, vmid, c.String("node"))
			default:
				return nil
			}
		},
	}
}
