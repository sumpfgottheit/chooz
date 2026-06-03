package main

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
		themeSlug      string
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

  theme:
    name: nord                           # optional; same as --theme nord

All unknown keys are silently ignored, so newer YAML files work with older
binaries without error.

THEMES

  Select a theme with --theme <slug> or CHOOZ_THEME=<slug>.
  Run 'chooz themes' to list all available themes.
  Run 'chooz theme-showroom' to browse themes interactively.

ENVIRONMENT

  NO_COLOR       disable all ANSI styling
  CLICOLOR_FORCE keep colours even when output is not a TTY
  CHOOZ_THEME    theme slug (overridden by --theme flag)

SHELL COMPLETION

  Activate for the current shell session:

    bash   source <(chooz completion bash)
    zsh    source <(chooz completion zsh)
    fish   chooz completion fish | source

  To persist across sessions, add the source line to ~/.bashrc or ~/.zshrc.

  Once active, tab-completion works for both arguments:
    chooz <TAB>            — completes .yaml / .yml files
    chooz menu.yaml <TAB>  — completes item names from that file

EXIT CODES

  0    item selected
  1    runtime error (bad YAML, name not found, no TTY without selection)
  2    usage error (bad flags or missing argument)
  130  user cancelled (Esc, Ctrl-C, or q)`,
		Version:       version,
		Args:          cobra.RangeArgs(0, 2),
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.Flags().StringVarP(&defaultName, "default", "d", "", "pre-highlight / CI fallback item name")
	root.Flags().StringVar(&themeSlug, "theme", "", "color theme slug (default: gum)")
	root.Flags().IntVar(&height, "height", 0, "max visible list rows")
	root.Flags().BoolVarP(&listFlag, "list", "l", false, "print all names and exit")
	root.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "never draw TUI")
	root.Flags().BoolVar(&noColor, "no-color", false, "disable all styling")

	_ = root.RegisterFlagCompletionFunc("theme", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return theme.Slugs(), cobra.ShellCompDirectiveNoFileComp
	})

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
		if len(args) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), strings.SplitN(cmd.Long, "\n", 2)[0])
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'chooz --help' for full documentation including YAML structure.")
			return nil
		}
		exitCode = dispatch(args, defaultName, themeSlug, height, listFlag, nonInteractive, noColor)
		return nil
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "themes",
			Short: "List all available themes",
			Long:  "List all built-in themes with their slug, name, and variant (dark/light/adaptive).",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				for _, p := range theme.All() {
					fmt.Fprintf(cmd.OutOrStdout(), "%-18s  %-30s  %s\n", p.Slug, p.Name, p.Variant)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "theme-showroom",
			Short: "Browse all built-in themes interactively",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				if err := tui.RunShowroom(); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
					exitCode = 1
				}
				return nil
			},
		},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	return exitCode
}

func dispatch(args []string, defaultName, themeSlug string, height int, listFlag, nonInteractive, noColor bool) int {
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
	// /dev/tty accessible → TUI can render there; stdout may still be piped.
	// Keep the file open so the Lipgloss renderer can query its color profile
	// throughout the TUI session (e.g. S=$(chooz menu.yaml) still gets colors).
	if !isTTY {
		if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
			defer f.Close()
			isTTY = true
			lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(f))
		}
	}

	// non-interactive path: --non-interactive flag or no usable TTY
	if nonInteractive || !isTTY {
		if defaultName != "" {
			fmt.Println(defaultName)
			return 0
		}
		fmt.Fprintln(os.Stderr, "error: not a terminal; pass a name or --default")
		return 1
	}

	// Resolve theme: --theme > CHOOZ_THEME env > YAML > gum (default)
	resolvedTheme := cmp.Or(themeSlug, os.Getenv("CHOOZ_THEME"), menu.Theme.Name)

	thm := theme.New(resolvedTheme, noColor)
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
