package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/lz-wang/pvectl/internal/output"
	"github.com/lz-wang/pvectl/internal/pve"
)

func newVMCommand(deps Dependencies) *cli.Command {
	return newGuestCommand("vm", "Manage QEMU virtual machines", deps)
}

func newLXCCommand(deps Dependencies) *cli.Command {
	return newGuestCommand("lxc", "Manage LXC containers", deps)
}

func newGuestCommand(kind, usage string, deps Dependencies) *cli.Command {
	return &cli.Command{
		Name:  kind,
		Usage: usage,
		Subcommands: []*cli.Command{
			{
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
			},
			{
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
			},
			guestControlCommand(kind, "start", "Start a guest", deps),
			guestControlCommand(kind, "shutdown", "Shutdown a guest gracefully", deps),
			guestControlCommand(kind, "stop", "Stop a guest immediately", deps),
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
			default:
				return nil
			}
		},
	}
}

func guestService(kind string, rt *runtime) *pve.GuestService {
	if kind == "vm" {
		return pve.NewVMService(rt.backend, rt.tasks, rt.logger, rt.verbose)
	}
	return pve.NewLXCService(rt.backend, rt.tasks, rt.logger, rt.verbose)
}
