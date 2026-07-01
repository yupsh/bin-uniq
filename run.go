package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-uniq"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "uniq"

const (
	flagCount      = "count"
	flagRepeated   = "repeated"
	flagUnique     = "unique"
	flagIgnoreCase = "ignore-case"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `uniq [OPTIONS] [FILE...]

Filter adjacent matching lines from FILE (or standard input),
writing one copy of each group to standard output.`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags (e.g. grep -v) while
// still exposing the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the uniq CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newApp(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newApp(version string, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "report or omit repeated lines",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: flagCount, Aliases: []string{"c"}, Usage: "prefix lines by the number of occurrences"},
			&cli.BoolFlag{
				Name:    flagRepeated,
				Aliases: []string{"d"},
				Usage:   "only print duplicate lines, one for each group",
			},
			&cli.BoolFlag{Name: flagUnique, Aliases: []string{"u"}, Usage: "only print unique lines"},
			&cli.BoolFlag{
				Name:    flagIgnoreCase,
				Aliases: []string{"i"},
				Usage:   "ignore differences in case when comparing",
			},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Uniq(options(c)...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

func options(c *cli.Command) []any {
	var opts []any
	if c.Bool(flagCount) {
		opts = append(opts, command.UniqCount)
	}
	if c.Bool(flagRepeated) {
		opts = append(opts, command.UniqDuplicatesOnly)
	}
	if c.Bool(flagUnique) {
		opts = append(opts, command.UniqUniqueOnly)
	}
	if c.Bool(flagIgnoreCase) {
		opts = append(opts, command.UniqIgnoreCase)
	}
	return opts
}
