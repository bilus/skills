// Command dfdmetrics reports shared state and package spread in a leveled dfd design.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one invocation and returns its exit status.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("dfdmetrics", flag.ContinueOnError)
	flags.SetOutput(stderr)
	score := flags.Bool("score", false, "print only the score card as tab-separated values")
	flags.Usage = func() { complain(stderr, "usage: dfdmetrics [-score] stem.dfd") }
	if err := flags.Parse(args); err != nil || flags.NArg() != 1 {
		if err == nil {
			flags.Usage()
		}
		return 2
	}
	d, err := dfdmetrics.Load(flags.Arg(0))
	if err != nil {
		complain(stderr, err.Error())
		return 1
	}
	std, err := dfdmetrics.StdPackages()
	if err != nil {
		complain(stderr, err.Error())
		return 1
	}
	r := dfdmetrics.Analyze(d, std)
	write := r.Write
	if *score {
		write = r.WriteScore
	}
	if err := write(stdout); err != nil {
		complain(stderr, err.Error())
		return 1
	}
	return 0
}

// complain writes msg and a newline to stderr.
func complain(stderr io.Writer, msg string) {
	// A failed write to stderr leaves no channel to report it on.
	_, _ = fmt.Fprintln(stderr, msg)
}
