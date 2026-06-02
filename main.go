package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/saf/chooz/internal/config"
	"github.com/saf/chooz/internal/selector"
	"github.com/saf/chooz/internal/theme"
	"github.com/saf/chooz/internal/tui"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		defaultName    string
		height         int
		listFlag       bool
		nonInteractive bool
		noColor        bool
	)

	root := &cobra.Command{
		Use:   "chooz <file.yaml> [name]",
		Short: "Interactive terminal menu driven by a YAML file",
		Long: `chooz — interactive terminal menu driven by a YAML file.

Renders a selectable list with a description preview pane. On selection the
chosen item's name is printed to stdout, making it composable in shell scripts:

  RESULT=$(chooz menu.yaml)

YAML STRUCTURE

  title: "Optional menu header"          # shown at the top
  description: "Optional help text"      # shown below the header

  items:                                 # required, must be non-empty
    - name: prod                         # required; only [A-Za-z0-9]+
      title: "Production"               # optional display label (default: name)
      description: |                    # optional, multi-line supported
        Live cluster. Handle with care.

  theme:                                 # optional colour overrides
    cursor:      "212"                   # ANSI-256 index or hex (#ff87d7)
    selected:    "212"
    item:        "252"
    header:      "252"
    description: "246"
    border:      "238"
    help:        "240"

All unknown keys are silently ignored, so newer YAML files work with older
binaries without error.

ENVIRONMENT

  NO_COLOR               disable all ANSI styling
  CLICOLOR_FORCE         keep colours even when output is not a TTY
  CHOOZ_THEME_CURSOR     override cursor colour (same format as theme.cursor)
  CHOOZ_THEME_SELECTED   override selected-item colour
  CHOOZ_THEME_ITEM       override normal-item colour
  CHOOZ_THEME_HEADER     override header colour
  CHOOZ_THEME_DESCRIPTION override description colour
  CHOOZ_THEME_BORDER     override border/separator colour
  CHOOZ_THEME_HELP       override help-hint colour

EXIT CODES

  0    item selected
  1    runtime error (bad YAML, name not found, no TTY without selection)
  2    usage error (bad flags or missing argument)
  130  user cancelled (Esc, Ctrl-C, or q)`,
		Version:       version,
		Args:          cobra.RangeArgs(1, 2),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.Flags().StringVarP(&defaultName, "default", "d", "", "pre-highlight / CI fallback item name")
	root.Flags().IntVar(&height, "height", 0, "max visible list rows")
	root.Flags().BoolVarP(&listFlag, "list", "l", false, "print all names and exit")
	root.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "never draw TUI")
	root.Flags().BoolVar(&noColor, "no-color", false, "disable all styling")

	root.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
		case 1:
			menu, err := config.Load(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return selector.List(menu.Items), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	exitCode := 0
	root.RunE = func(cmd *cobra.Command, args []string) error {
		exitCode = dispatch(args, defaultName, height, listFlag, nonInteractive, noColor)
		return nil
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	return exitCode
}

func dispatch(args []string, defaultName string, height int, listFlag, nonInteractive, noColor bool) int {
	filePath := args[0]
	var posName string
	if len(args) > 1 {
		posName = args[1]
	}

	menu, err := config.Load(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// validate --default early so errors are consistent regardless of mode
	if defaultName != "" {
		if _, err := selector.Resolve(menu.Items, defaultName); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
	}

	// --list: print all names and exit
	if listFlag {
		for _, name := range selector.List(menu.Items) {
			fmt.Println(name)
		}
		return 0
	}

	// positional name: resolve and print without TUI
	if posName != "" {
		if _, err := selector.Resolve(menu.Items, posName); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println(posName)
		return 0
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	// non-interactive path: --non-interactive flag or stdout not a TTY
	if nonInteractive || !isTTY {
		if defaultName != "" {
			fmt.Println(defaultName)
			return 0
		}
		fmt.Fprintln(os.Stderr, "error: not a terminal; pass a name or --default")
		return 1
	}

	// interactive TUI
	thm := theme.New(menu.Theme, noColor)
	result, err := tui.Run(menu, thm, defaultName, height)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if result == "" {
		return 130
	}
	fmt.Println(result)
	return 0
}
