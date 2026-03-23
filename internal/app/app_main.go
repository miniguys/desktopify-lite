package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func Main() {
	if err := run(os.Stdin); err != nil {
		fmt.Println(styleError.Render(" " + err.Error()))
		os.Exit(1)
	}
}

func run(stdin io.Reader) error {
	runtimeOpts, err := ParseRuntimeOptions(osArgs())
	if err != nil {
		return err
	}

	switch runtimeOpts.Command {
	case CommandHelp:
		printHelp()
		return nil
	case CommandVersion:
		printVersion()
		return nil
	case CommandConfig:
		return runConfigCommand(runtimeOpts)
	case CommandConfigReset:
		return runConfigResetCommand()
	}

	cfg, meta, err := LoadConfig()
	if err != nil {
		return err
	}

	if err := ValidateProxyURL(cfg.DefaultProxy); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	effectiveProxy := EffectiveProxy(runtimeOpts, cfg)

	if !runtimeOpts.NonInteractive {
		printLogo()
	}

	if cfg.WithDebug {
		fmt.Println(styleInfoName.Render("Config:"), styleInfoValue.Render(meta.ActivePath))
		if meta.CreatedDefaultPath != "" {
			fmt.Println(styleInfoName.Render("Created default config:"), styleInfoValue.Render(meta.CreatedDefaultPath))
		}
		if meta.CreatedExamplePath != "" {
			fmt.Println(styleInfoName.Render("Created example config:"), styleInfoValue.Render(meta.CreatedExamplePath))
		}
		if effectiveProxy != "" {
			fmt.Println(styleInfoName.Render("Proxy:"), styleInfoValue.Render(effectiveProxy))
		}
		fmt.Println()
	}

	if !runtimeOpts.NonInteractive {
		printInfoBlock()
	}

	in, err := ResolveRunInput(runtimeOpts, cfg, bufio.NewReader(stdin))
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Proxy) != "" {
		effectiveProxy = strings.TrimSpace(in.Proxy)
	}

	paths, err := DefaultPaths()
	if err != nil {
		return err
	}

	if err := EnsureDirs(paths); err != nil {
		return err
	}

	iconRef := ""
	if strings.TrimSpace(in.IconURL) != "" {
		iconRef, err = ResolveIcon(in, paths, cfg, effectiveProxy)
		if err != nil && !in.IconURLExplicit {
			if runtimeOpts.NonInteractive {
				fmt.Fprintln(os.Stderr, styleWarning.Render(" Warning:"), "icon auto-discovery failed for", in.URL, err)
				fmt.Fprintln(os.Stderr, styleWarning.Render(" Warning:"), "continuing without an icon")
			} else {
				fmt.Println(styleError.Render(fmt.Sprintf("%v", err)))
				fmt.Println(styleInfoName.Render(" Skipping icon..."))
			}
			iconRef = ""
		} else if err != nil {
			return fmt.Errorf("resolve explicit icon url: %w", err)
		}
	}

	entry, filename, err := BuildDesktopEntry(in, iconRef, cfg)
	if err != nil {
		return err
	}

	desktopPath := paths.DesktopFilePath(filename)
	if err := WriteDesktopFile(desktopPath, entry); err != nil {
		return err
	}
	if iconRef != "" {
		if err := removeStaleIconVariants(paths, in.Name, iconRef); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cleanup stale icon variants: %v\n", err)
		}
	}

	fmt.Println(styleSuccess.Render("✔  Success! App '" + in.Name + "' is ready."))
	fmt.Println(styleInfoName.Render(" Path: "), styleInfoValue.Render(desktopPath))
	fmt.Println()

	if cfg.WithDebug {
		fmt.Println(styleProcess.Render("Tip: If it doesn't show up, try: update-desktop-database " + paths.ApplicationsDir))
		fmt.Println()
	}

	return nil
}

func runConfigCommand(runtimeOpts RuntimeOptions) error {
	cfg, meta, err := LoadConfig()
	repairingBrokenConfig := false
	if err != nil {
		if strings.TrimSpace(meta.ActivePath) == "" {
			return err
		}
		cfg = DefaultConfig()
		repairingBrokenConfig = true
	}

	targetPath, err := ConfigTargetPath(meta)
	if err != nil {
		return err
	}

	updated, err := ApplyConfigUpdates(cfg, runtimeOpts.ConfigUpdates)
	if err != nil {
		return err
	}
	if err := WriteConfigFile(targetPath, updated); err != nil {
		return err
	}

	if repairingBrokenConfig {
		fmt.Fprintf(os.Stderr, "warning: repaired invalid config at %s using defaults plus requested updates\n", targetPath)
	}
	fmt.Println(styleSuccess.Render(" Config updated."))
	fmt.Println(styleInfoName.Render(" Path: "), styleInfoValue.Render(targetPath))
	return nil
}

func runConfigResetCommand() error {
	targetPath, err := ConfigTargetPath(ConfigLoadMeta{})
	if err != nil {
		return err
	}
	if err := WriteConfigFile(targetPath, DefaultConfig()); err != nil {
		return err
	}

	fmt.Println(styleSuccess.Render(" Config reset to defaults."))
	fmt.Println(styleInfoName.Render(" Path: "), styleInfoValue.Render(targetPath))
	return nil
}

func printHelp() {
	fmt.Println("desktopify-lite")
	fmt.Println("Generate a Linux .desktop launcher for a website.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  desktopify-lite")
	fmt.Println("  desktopify-lite --url=https://example.com --name='Example App' [--icon-url=https://example.com/icon.png|file:///path/icon.svg|./icon.png] [--skip-icon] [--browser=chromium] [--url-template='--app={url}'] [--extra-flags='--incognito'] [--startup-wm-class=ExampleApp] [--proxy=http://127.0.0.1:8080]")
	fmt.Println("  desktopify-lite config [--default_browser=chromium] [--default_url_template=--app={url}] [--default_extra_flags=--incognito] [--default_proxy=http://127.0.0.1:8080] [--disable_google_favicon=true] [--with_debug=true]")
	fmt.Println("  desktopify-lite config-reset")
	fmt.Println("  desktopify-lite --version")
	fmt.Println("  desktopify-lite version")
	fmt.Println("  desktopify-lite help")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  config        Update config values in the active config file.")
	fmt.Println("  config-reset  Replace the active config file with factory defaults.")
	fmt.Println("  version       Show version.")
	fmt.Println("  help          Show this help text.")
	fmt.Println()
	fmt.Println("Run flags:")
	fmt.Println("  --url                 Website URL. Required in non-interactive mode.")
	fmt.Println("  --name                Launcher name. Required in non-interactive mode.")
	fmt.Println("  --icon-url            Icon URL or local icon path. Defaults to the website URL for auto-discovery.")
	fmt.Println("  --skip-icon, --no-icon Skip icon resolution entirely.")
	fmt.Println("  --browser             Browser binary.")
	fmt.Println("  --url-template        URL template fragment. Must contain {url}; may expand to multiple argv parts.")
	fmt.Println("  --extra-flags         Extra browser flags.")
	fmt.Println("  --startup-wm-class    Optional StartupWMClass value.")
	fmt.Println("  --proxy               Proxy URL for HTTP requests made during the current run.")
	fmt.Println("  --help, -h            Show help.")
	fmt.Println("  --version, -v         Show version.")
	fmt.Println()
	fmt.Println("Config flags:")
	fmt.Println("  --default_browser      Default browser binary used when the prompt value is empty.")
	fmt.Println("  --default_url_template URL template. Must contain {url}.")
	fmt.Println("  --default_extra_flags  Extra browser flags used when the prompt value is empty.")
	fmt.Println("  --default_proxy        Default proxy URL used when --proxy is not provided.")
	fmt.Println("  --disable_google_favicon Disable Google favicon lookup for icon auto-discovery.")
	fmt.Println("  --with_debug           Enable or disable debug output in the config file.")
}

var (
	errCanceled = errors.New("canceled")
	osArgs      = func() []string { return os.Args[1:] }
)
