package check

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type CheckStruct struct {
	Name   string
	Option *pflag.FlagSet
}

func New(name string) *CheckStruct {
	check := &CheckStruct{
		Name:   name,
		Option: pflag.NewFlagSet(name, 1),
	}

	return check
}

func (c CheckStruct) Init() {
	c.Option.Parse(os.Args[1:])
}

func (c CheckStruct) Ok(output string) {
	fmt.Println(c.Name, "OK:", output)
	os.Exit(0)
}

func (c CheckStruct) Warning(output string) {
	fmt.Println(c.Name, "WARNING:", output)
	os.Exit(1)
}

func (c CheckStruct) Critical(output string) {
	fmt.Println(c.Name, "CRITICAL:", output)
	os.Exit(2)
}

func (c CheckStruct) Error(err error) {
	fmt.Println(c.Name, "ERROR:", err)
	os.Exit(3)
}
