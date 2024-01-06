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

type session struct {
	Input input
	FstypeExcludes []string
	MountExcludes  []string
	FWarn  float64
	FCrit  float64
	WarnMnt []string
	CritMnt []string
	Perf    []string
	Perfs   string
}

func main() {
	var session session

	c := check.New("CheckDisk")
	c.Option.IntVarP(&session.Input.Warn, "warn", "w", 80, "Warning percentage (greater than or equal to) threshold")
	c.Option.IntVarP(&session.Input.Crit, "crit", "c", 100, "Critical percentage (greater than or equal to) threshold")
	c.Option.Float64VarP(&session.Input.Magic, "magic", "m", 1.0, "Magic factor to adjust thresholds.  Example: 0.9")
	c.Option.Float64VarP(&session.Input.Normal, "normal", "n", 20, "\"Normal\" size in GB, thresholds are not adjusted for filesystems of exactly this size, levels are reduced for smaller file systems and raised for larger filesystems")
	c.Option.Float64VarP(&session.Input.Minimum, "minimum", "l", 100, "Minimum size in GB, before applying magic adjustment")
	c.Option.StringVarP(&session.Input.FstypeExclude, "exclude", "x", "", "Comma separated list of file system types to exclude")
	c.Option.StringVarP(&session.Input.MountExclude, "ignore", "i", "", "Comma separated list of mount points to ignore")
	c.Option.StringVarP(&session.Input.Path, "path", "p", "", "Limit check to specified path")
	c.Init()

	(&session).parseExcludes()

	usage, err := diskUsage(session.Input.Path)
	if err != nil {
		c.Error(err)
	}

	for _, u := range usage {
		if !Contains(session.FstypeExcludes, u[1]) && !Contains(session.MountExcludes, u[6]) {
			cap, err := strconv.ParseFloat(strings.TrimRight(u[5], "%"), 64)
			if err != nil {
				c.Error(err)
			}

			f_size, err := strconv.ParseFloat(u[2], 64)
			if err != nil {
				c.Error(err)
			}

			if f_size*1024 >= session.Input.Minimum*1073741824 {
				session.FCrit = adjPercent(f_size, float64(session.Input.Crit), session.Input.Magic, session.Input.Normal)
				session.FWarn = adjPercent(f_size, float64(session.Input.Warn), session.Input.Magic, session.Input.Normal)
			} else {
				session.FCrit = float64(session.Input.Crit)
				session.FWarn = float64(session.Input.Warn)
			}
			switch {
			case cap >= session.FCrit:
				session.CritMnt = append(session.CritMnt, u[6]+" "+u[5])
			case cap >= session.FWarn:
				session.WarnMnt = append(session.WarnMnt, u[6]+" "+u[5])
			}
			session.Perf = append(session.Perf, fmt.Sprintf("%s=%s;%.2f;%.2f", u[6], u[5], session.FWarn, session.FCrit))
		}
	}

	session.Perfs = strings.Join(session.Perf, " ")
	switch {
	case len(session.CritMnt) > 0:
		c.Critical(strings.Join(session.CritMnt, ", ") + " | " + session.Perfs)
	case len(session.WarnMnt) > 0:
		c.Warning(strings.Join(session.WarnMnt, ", ") + " | " + session.Perfs)
	default:
		c.Ok("OK" + " | " + session.Perfs)
	}
}

func (s *session) parseExcludes() {
	s.FstypeExcludes = strings.Split(s.Input.FstypeExclude, ",")
	s.MountExcludes = strings.Split(s.Input.MountExclude, ",")
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
