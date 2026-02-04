package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

const (
	colorAuto  = "auto"
	colorNever = "never"
)

func helpOptions() kong.HelpOptions {
	return kong.HelpOptions{
		NoExpandSubcommands: true,
	}
}

func helpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	origStdout := ctx.Stdout

	width := guessColumns(origStdout)

	oldCols, hadCols := os.LookupEnv("COLUMNS")
	_ = os.Setenv("COLUMNS", strconv.Itoa(width))

	defer func() {
		if hadCols {
			_ = os.Setenv("COLUMNS", oldCols)
		} else {
			_ = os.Unsetenv("COLUMNS")
		}
	}()

	buf := bytes.NewBuffer(nil)
	ctx.Stdout = buf

	defer func() { ctx.Stdout = origStdout }()

	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}

	out := buf.String()
	out = injectBuildLine(out)
	out = colorizeHelp(out, helpProfile(origStdout, helpColorMode(ctx.Args)))
	_, err := io.WriteString(origStdout, out)

	return err
}

func injectBuildLine(out string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}

	c := strings.TrimSpace(commit)
	line := fmt.Sprintf("Build: %s", v)

	if c != "" {
		line = fmt.Sprintf("%s (%s)", line, c)
	}

	lines := strings.Split(out, "\n")

	for i, l := range lines {
		if strings.HasPrefix(l, "Usage:") {
			if i+1 < len(lines) && lines[i+1] == line {
				return out
			}

			outLines := make([]string, 0, len(lines)+1)
			outLines = append(outLines, lines[:i+1]...)
			outLines = append(outLines, line)
			outLines = append(outLines, lines[i+1:]...)

			return strings.Join(outLines, "\n")
		}
	}

	return out
}

func helpColorMode(args []string) string {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("PONTO_COLOR"))); v != "" {
		return v
	}

	for i := range len(args) {
		a := args[i]
		if a == "--plain" || a == "--json" || a == "--csv" {
			return colorNever
		}

		if a == "--color" && i+1 < len(args) {
			return strings.ToLower(strings.TrimSpace(args[i+1]))
		}

		if v, ok := strings.CutPrefix(a, "--color="); ok {
			return strings.ToLower(strings.TrimSpace(v))
		}
	}

	return colorAuto
}

func helpProfile(stdout io.Writer, mode string) termenv.Profile {
	if termenv.EnvNoColor() {
		return termenv.Ascii
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = colorAuto
	}

	switch mode {
	case colorNever:
		return termenv.Ascii
	case "always":
		return termenv.TrueColor
	default:
		o := termenv.NewOutput(stdout, termenv.WithProfile(termenv.EnvColorProfile()))
		return o.Profile
	}
}

func colorizeHelp(out string, profile termenv.Profile) string {
	if profile == termenv.Ascii {
		return out
	}

	heading := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#60a5fa")).Bold().String()
	}
	section := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#a78bfa")).Bold().String()
	}
	cmdName := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#38bdf8")).Bold().String()
	}
	dim := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#9ca3af")).String()
	}

	inCommands := false
	lines := strings.Split(out, "\n")

	for i, line := range lines {
		if line == "Commands:" {
			inCommands = true
		}

		switch {
		case strings.HasPrefix(line, "Usage:"):
			lines[i] = heading("Usage:") + strings.TrimPrefix(line, "Usage:")
		case line == "Flags:":
			lines[i] = section(line)
		case line == "Commands:":
			lines[i] = section(line)
		case line == "Arguments:":
			lines[i] = section(line)
		case strings.HasPrefix(line, "Build:"):
			lines[i] = section(line)
		case inCommands && strings.HasPrefix(line, "  ") && (len(line) < 3 || line[2] != ' '):
			lines[i] = colorizeCommandLine(line, cmdName, dim)
		case inCommands && strings.HasPrefix(line, "    ") && strings.TrimSpace(line) != "":
			lines[i] = "    " + dim(strings.TrimPrefix(line, "    "))
		}
	}

	return strings.Join(lines, "\n")
}

func colorizeCommandLine(line string, cmdName func(string) string, dim func(string) string) string {
	if !strings.HasPrefix(line, "  ") {
		return line
	}

	rest := strings.TrimPrefix(line, "  ")
	if rest == "" {
		return line
	}

	name, tail, _ := strings.Cut(rest, " ")
	if name == "" {
		return line
	}

	styled := cmdName(name)
	if tail == "" {
		return "  " + styled
	}

	tail = strings.ReplaceAll(tail, "<", dim("<"))
	tail = strings.ReplaceAll(tail, ">", dim(">"))
	tail = strings.ReplaceAll(tail, "[flags]", dim("[flags]"))

	return "  " + styled + " " + tail
}

func guessColumns(w io.Writer) int {
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if cols, err := strconv.Atoi(colsStr); err == nil {
			return cols
		}
	}

	f, ok := w.(*os.File)
	if !ok {
		return 80
	}

	width, _, err := term.GetSize(int(f.Fd()))
	if err == nil && width > 0 {
		return width
	}

	return 80
}
