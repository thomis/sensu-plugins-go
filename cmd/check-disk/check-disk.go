package main

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"github.com/thomis/sensu-plugins-go/pkg/check"
)

type input struct {
	Warn          int
	Crit          int
	Normal        float64
	Magic         float64
	Minimum       float64
	FstypeExclude string
	MountExclude  string
	Path          string
}

func main() {
	var (
		input input
		fstype_excludes []string
		mount_excludes  []string
		f_warn  float64
		f_crit  float64
		warnMnt []string
		critMnt []string
		perf    []string
		perfs   string
	)

	c := check.New("CheckDisk")
	c.Option.IntVarP(&input.Warn, "warn", "w", 80, "Warning percentage (greater than or equal to) threshold")
	c.Option.IntVarP(&input.Crit, "crit", "c", 100, "Critical percentage (greater than or equal to) threshold")
	c.Option.Float64VarP(&input.Magic, "magic", "m", 1.0, "Magic factor to adjust thresholds.  Example: 0.9")
	c.Option.Float64VarP(&input.Normal, "normal", "n", 20, "\"Normal\" size in GB, thresholds are not adjusted for filesystems of exactly this size, levels are reduced for smaller file systems and raised for larger filesystems")
	c.Option.Float64VarP(&input.Minimum, "minimum", "l", 100, "Minimum size in GB, before applying magic adjustment")
	c.Option.StringVarP(&input.FstypeExclude, "exclude", "x", "", "Comma separated list of file system types to exclude")
	c.Option.StringVarP(&input.MountExclude, "ignore", "i", "", "Comma separated list of mount points to ignore")
	c.Option.StringVarP(&input.Path, "path", "p", "", "Limit check to specified path")
	c.Init()

	fstype_excludes = strings.Split(input.FstypeExclude, ",")
	mount_excludes = strings.Split(input.MountExclude, ",")

	usage, err := diskUsage(input.Path)
	if err != nil {
		c.Error(err)
	}

	for _, u := range usage {
		if !Contains(fstype_excludes, u[1]) && !Contains(mount_excludes, u[6]) {
			cap, err := strconv.ParseFloat(strings.TrimRight(u[5], "%"), 64)
			if err != nil {
				c.Error(err)
			}

			f_size, err := strconv.ParseFloat(u[2], 64)
			if err != nil {
				c.Error(err)
			}

			if f_size*1024 >= input.Minimum*1073741824 {
				f_crit = adjPercent(f_size, float64(input.Crit), input.Magic, input.Normal)
				f_warn = adjPercent(f_size, float64(input.Warn), input.Magic, input.Normal)
			} else {
				f_crit = float64(input.Crit)
				f_warn = float64(input.Warn)
			}
			switch {
			case cap >= f_crit:
				critMnt = append(critMnt, u[6]+" "+u[5])
			case cap >= f_warn:
				warnMnt = append(warnMnt, u[6]+" "+u[5])
			}
			perf = append(perf, fmt.Sprintf("%s=%s;%.2f;%.2f", u[6], u[5], f_warn, f_crit))
		}
	}

	perfs = strings.Join(perf, " ")
	switch {
	case len(critMnt) > 0:
		c.Critical(strings.Join(critMnt, ", ") + " | " + perfs)
	case len(warnMnt) > 0:
		c.Warning(strings.Join(warnMnt, ", ") + " | " + perfs)
	default:
		c.Ok("OK" + " | " + perfs)
	}
}

func diskUsage(path string) ([][]string, error) {
	var (
		out []byte
		err error
	)
	if len(path) > 0 {
		out, err = exec.Command("df", "-lTP", path).Output()
	} else {
		out, err = exec.Command("df", "-lTP").Output()
	}
	if err != nil {
		return [][]string{}, err
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")[1:]
	result := make([][]string, len(lines))

	for i := range lines {
		stats := strings.Fields(lines[i])
		// return device, fstype, size, used, avail, pctused, mountpoint
		result[i] = []string{stats[0], stats[1], stats[2], stats[3], stats[4], stats[5], stats[6]}
	}

	return result, nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func adjPercent(size float64, percent float64, magic float64, normal float64) float64 {
	hsize := (size / (1024.0 * 1024.0)) / normal
	felt := math.Pow(hsize, magic)
	scale := felt / hsize
	return 100 - ((100 - percent) * scale)
}
