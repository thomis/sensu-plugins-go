package main

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"github.com/portertech/sensu-plugins-go/lib/check"
)

func main() {
	var (
		warn            int
		crit            int
		fstype_exclude  string
		fstype_excludes []string
		mount_exclude   string
		mount_excludes  []string
		normal          float64
		magic           float64
		minimum         float64
		path            string
		f_warn          float64
		f_crit          float64
		warnMnt         []string
		critMnt         []string
		perf            []string
		perfs           string
	)

	c := check.New("CheckDisk")
	c.Option.IntVarP(&warn, "warn", "w", 80, "Warning percentage (greater than or equal to) threshold")
	c.Option.IntVarP(&crit, "crit", "c", 100, "Critical percentage (greater than or equal to) threshold")
	c.Option.Float64VarP(&magic, "magic", "m", 1.0, "Magic factor to adjust thresholds.  Example: 0.9")
	c.Option.Float64VarP(&normal, "normal", "n", 20, "\"Normal\" size in GB, thresholds are not adjusted for filesystems of exactly this size, levels are reduced for smaller file systems and raised for larger filesystems")
	c.Option.Float64VarP(&minimum, "minimum", "l", 100, "Minimum size in GB, before applying magic adjustment")
	c.Option.StringVarP(&fstype_exclude, "exclude", "x", "", "Comma separated list of file system types to exclude")
	c.Option.StringVarP(&mount_exclude, "ignore", "i", "", "Comma separated list of mount points to ignore")
	c.Option.StringVarP(&path, "path", "p", "", "Limit check to specified path")
	c.Init()

	fstype_excludes = strings.Split(fstype_exclude, ",")
	mount_excludes = strings.Split(mount_exclude, ",")

	usage, err := diskUsage(path)
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

			if f_size*1024 >= minimum*1073741824 {
				f_crit = adj_percent(f_size, float64(crit), magic, normal)
				f_warn = adj_percent(f_size, float64(warn), magic, normal)
			} else {
				f_crit = float64(crit)
				f_warn = float64(warn)
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
func adj_percent(size float64, percent float64, magic float64, normal float64) float64 {
	hsize := (size / (1024.0 * 1024.0)) / normal
	felt := math.Pow(hsize, magic)
	scale := felt / hsize
	return 100 - ((100 - percent) * scale)
}
