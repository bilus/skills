// Command dfdreview builds the review page of a leveled dfd design.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run executes one invocation and returns its exit status.
func run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("dfdreview", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var opts dfdreview.Options
	flags.StringVar(&opts.Base, "base", "", "revision to compare against, such as the stage's start commit")
	flags.StringVar(&opts.Vocabulary, "vocabulary", "", "the vocabulary (default: vocabulary.md beside the design)")
	flags.StringVar(&opts.Plan, "plan", "", "the plan, for its sections \"The change in brief\" and \"Metaphor\"")
	flags.StringVar(&opts.OldScore, "old-score", "", "the cached score card of the last approved review")
	flags.StringVar(&opts.NewScore, "new-score", "", "this review's score card, from dfdmetrics -score")
	flags.StringVar(&opts.Report, "report", "", "this review's report, from dfdmetrics")
	flags.StringVar(&opts.DFD, "dfd", "dfd", "the dfd command")
	flags.StringVar(&opts.Out, "o", "", "the page (default: review/index.html beside the design)")
	flags.Usage = func() {
		complain(stderr, "usage: dfdreview [flags] docs/flow.dfd")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil || flags.NArg() != 1 {
		if err == nil {
			flags.Usage()
		}
		return 2
	}
	opts.Design = flags.Arg(0)
	dir := filepath.Dir(opts.Design)
	if opts.Vocabulary == "" {
		opts.Vocabulary = filepath.Join(dir, "vocabulary.md")
	}
	if opts.Out == "" {
		opts.Out = filepath.Join(dir, "review", "index.html")
	}
	if err := dfdreview.Build(opts); err != nil {
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
