package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

type NodeRow struct {
	Name    string  `json:"name" yaml:"name"`
	Status  string  `json:"status" yaml:"status"`
	CPU     float64 `json:"cpu" yaml:"cpu"`
	Mem     uint64  `json:"mem" yaml:"mem"`
	MaxMem  uint64  `json:"max_mem" yaml:"max_mem"`
	Disk    uint64  `json:"disk" yaml:"disk"`
	MaxDisk uint64  `json:"max_disk" yaml:"max_disk"`
	Uptime  uint64  `json:"uptime" yaml:"uptime"`
}

type GuestRow struct {
	Kind    string  `json:"kind" yaml:"kind"`
	VMID    uint64  `json:"vmid" yaml:"vmid"`
	Name    string  `json:"name" yaml:"name"`
	Node    string  `json:"node" yaml:"node"`
	Status  string  `json:"status" yaml:"status"`
	CPUs    int     `json:"cpus" yaml:"cpus"`
	CPU     float64 `json:"cpu" yaml:"cpu"`
	Mem     uint64  `json:"mem" yaml:"mem"`
	MaxMem  uint64  `json:"max_mem" yaml:"max_mem"`
	MaxDisk uint64  `json:"max_disk" yaml:"max_disk"`
	Uptime  uint64  `json:"uptime" yaml:"uptime"`
	Tags    string  `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func ValidateFormat(format string) error {
	switch strings.ToLower(format) {
	case FormatTable, FormatJSON, FormatYAML:
		return nil
	default:
		return fmt.Errorf("invalid output format %q, expected table, json, or yaml", format)
	}
}

func NormalizeFormat(format string) string {
	if format == "" {
		return FormatTable
	}
	return strings.ToLower(format)
}

func Write(w io.Writer, format string, value any, writeTable func(io.Writer) error) error {
	format = NormalizeFormat(format)
	if err := ValidateFormat(format); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	case FormatYAML:
		data, err := yaml.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode yaml: %w", err)
		}
		_, err = w.Write(data)
		return err
	case FormatTable:
		if writeTable == nil {
			return errors.New("table renderer is required")
		}
		return writeTable(w)
	default:
		return nil
	}
}

func WriteNodeRows(w io.Writer, format string, rows []NodeRow) error {
	return Write(w, format, rows, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw, "NAME\tSTATUS\tCPU\tMEM\tMAXMEM\tDISK\tMAXDISK\tUPTIME"); err != nil {
			return err
		}
		for _, row := range rows {
			if _, err := fmt.Fprintf(
				tw,
				"%s\t%s\t%.2f\t%s\t%s\t%s\t%s\t%s\n",
				row.Name,
				row.Status,
				row.CPU,
				FormatBytes(row.Mem),
				FormatBytes(row.MaxMem),
				FormatBytes(row.Disk),
				FormatBytes(row.MaxDisk),
				FormatUptime(row.Uptime),
			); err != nil {
				return err
			}
		}
		return tw.Flush()
	})
}

func WriteGuestRows(w io.Writer, format string, rows []GuestRow) error {
	return Write(w, format, rows, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw, "VMID\tNAME\tNODE\tSTATUS\tCPUS\tMEM\tMAXMEM\tMAXDISK\tUPTIME\tTAGS"); err != nil {
			return err
		}
		for _, row := range rows {
			if _, err := fmt.Fprintf(
				tw,
				"%d\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
				row.VMID,
				empty(row.Name),
				row.Node,
				row.Status,
				row.CPUs,
				FormatBytes(row.Mem),
				FormatBytes(row.MaxMem),
				FormatBytes(row.MaxDisk),
				FormatUptime(row.Uptime),
				empty(row.Tags),
			); err != nil {
				return err
			}
		}
		return tw.Flush()
	})
}

func WriteGuestDetail(w io.Writer, format string, row GuestRow) error {
	return Write(w, format, row, func(w io.Writer) error {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		lines := [][2]string{
			{"Kind", row.Kind},
			{"VMID", fmt.Sprint(row.VMID)},
			{"Name", empty(row.Name)},
			{"Node", row.Node},
			{"Status", row.Status},
			{"CPUs", fmt.Sprint(row.CPUs)},
			{"CPU", fmt.Sprintf("%.2f", row.CPU)},
			{"Memory", FormatBytes(row.Mem)},
			{"Max Memory", FormatBytes(row.MaxMem)},
			{"Max Disk", FormatBytes(row.MaxDisk)},
			{"Uptime", FormatUptime(row.Uptime)},
			{"Tags", empty(row.Tags)},
		}
		for _, line := range lines {
			if _, err := fmt.Fprintf(tw, "%s:\t%s\n", line[0], line[1]); err != nil {
				return err
			}
		}
		return tw.Flush()
	})
}

func FormatBytes(n uint64) string {
	if n == 0 {
		return "0B"
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	value := float64(n)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%dB", n)
	}
	return fmt.Sprintf("%.1f%s", value, units[unit])
}

func FormatUptime(seconds uint64) string {
	if seconds == 0 {
		return "-"
	}
	d := time.Duration(seconds) * time.Second
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	mins := d / time.Minute

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if mins > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", mins))
	}
	return strings.Join(parts, "")
}

func empty(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
