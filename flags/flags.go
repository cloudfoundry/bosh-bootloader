package flags

import (
	"flag"
	"io/ioutil"
)

type Flags struct {
	set *flag.FlagSet
}

func New(name string) Flags {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.Usage = func() {}
	set.SetOutput(ioutil.Discard)

	return Flags{
		set: set,
	}
}

func (f Flags) String(v *string, name string, value string) {
	f.set.StringVar(v, name, value, "")
}

func (f Flags) Bool(v *bool, name string) {
	f.set.BoolVar(v, name, false, "")
}

func (f Flags) Parse(args []string) error {
	return f.set.Parse(args)
}

func (f Flags) Args() []string {
	return f.set.Args()
}
