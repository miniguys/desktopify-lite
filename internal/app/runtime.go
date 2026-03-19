package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Command string

const (
	CommandRun         Command = "run"
	CommandConfig      Command = "config"
	CommandConfigReset Command = "config-reset"
	CommandHelp        Command = "help"
	CommandVersion     Command = "version"
)

type trackedString struct {
	value string
	set   bool
}

func (t *trackedString) String() string { return t.value }
func (t *trackedString) Set(v string) error {
	t.value = v
	t.set = true
	return nil
}

type trackedBool struct {
	value bool
	set   bool
}

func (t *trackedBool) String() string {
	if !t.value {
		return "false"
	}
	return "true"
}
func (t *trackedBool) Set(v string) error {
	parsed, err := parseBoolStrict(v)
	if err != nil {
		return err
	}
	t.value = parsed
	t.set = true
	return nil
}

type ConfigUpdates struct {
	DefaultBrowser       *string
	DefaultURLTemplate   *string
	DefaultExtraFlags    *string
	DefaultProxy         *string
	DisableGoogleFavicon *bool
	WithDebug            *bool
}

type RuntimeOptions struct {
	Command          Command
	Proxy            string
	NonInteractive   bool
	SkipIcon         bool
	RunInput         Input
	RunInputExplicit RunInputExplicit
	ConfigUpdates    ConfigUpdates
}

type RunInputExplicit struct {
	URL            bool
	Name           bool
	IconURL        bool
	Browser        bool
	URLTemplate    bool
	ExtraFlags     bool
	StartupWMClass bool
}

func ParseRuntimeOptions(args []string) (RuntimeOptions, error) {
	if len(args) == 0 {
		return RuntimeOptions{Command: CommandRun}, nil
	}

	switch args[0] {
	case "help":
		return RuntimeOptions{Command: CommandHelp}, nil
	case "version":
		return RuntimeOptions{Command: CommandVersion}, nil
	case "config":
		return parseConfigCommand(args[1:])
	case "config-reset":
		return parseConfigResetCommand(args[1:])
	}

	fs := flag.NewFlagSet("desktopify-lite", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var opts RuntimeOptions
	opts.Command = CommandRun
	var showHelp bool
	var showVersion bool
	var runURL trackedString
	var runName trackedString
	var runIconURL trackedString
	var runBrowser trackedString
	var runURLTemplate trackedString
	var runExtraFlags trackedString
	var runStartupWMClass trackedString
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.BoolVar(&showVersion, "v", false, "show version")
	fs.StringVar(&opts.Proxy, "proxy", "", "proxy URL for all HTTP requests")
	fs.Var(&runURL, "url", "website URL")
	fs.Var(&runName, "name", "launcher name")
	fs.Var(&runIconURL, "icon-url", "icon URL")
	fs.Var(&runBrowser, "browser", "browser binary")
	fs.Var(&runURLTemplate, "url-template", "URL template")
	fs.Var(&runExtraFlags, "extra-flags", "extra browser flags")
	fs.Var(&runStartupWMClass, "startup-wm-class", "StartupWMClass value")
	fs.BoolVar(&opts.SkipIcon, "skip-icon", false, "skip icon resolution entirely")
	fs.BoolVar(&opts.SkipIcon, "no-icon", false, "skip icon resolution entirely")

	if err := fs.Parse(args); err != nil {
		return RuntimeOptions{}, err
	}
	if showHelp {
		return RuntimeOptions{Command: CommandHelp}, nil
	}
	if showVersion {
		return RuntimeOptions{Command: CommandVersion}, nil
	}
	if fs.NArg() != 0 {
		return RuntimeOptions{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}
	if opts.SkipIcon && runIconURL.set {
		return RuntimeOptions{}, errors.New("cannot use --icon-url with --skip-icon/--no-icon")
	}
	if err := ValidateProxyURL(opts.Proxy); err != nil {
		return RuntimeOptions{}, err
	}

	if runURL.set {
		opts.RunInput.URL = runURL.value
		opts.RunInputExplicit.URL = true
	}
	if runName.set {
		opts.RunInput.Name = runName.value
		opts.RunInputExplicit.Name = true
	}
	if runIconURL.set {
		opts.RunInput.IconURL = runIconURL.value
		opts.RunInputExplicit.IconURL = true
	}
	if runBrowser.set {
		opts.RunInput.Browser = runBrowser.value
		opts.RunInputExplicit.Browser = true
	}
	if runURLTemplate.set {
		opts.RunInput.URLTemplate = runURLTemplate.value
		opts.RunInputExplicit.URLTemplate = true
	}
	if runExtraFlags.set {
		opts.RunInput.ExtraFlags = runExtraFlags.value
		opts.RunInputExplicit.ExtraFlags = true
	}
	if runStartupWMClass.set {
		opts.RunInput.StartupWMClass = runStartupWMClass.value
		opts.RunInputExplicit.StartupWMClass = true
	}

	opts.NonInteractive = hasRunFlags(opts)
	return opts, nil
}

func hasRunFlags(opts RuntimeOptions) bool {
	return opts.SkipIcon ||
		strings.TrimSpace(opts.RunInput.URL) != "" ||
		strings.TrimSpace(opts.RunInput.Name) != "" ||
		strings.TrimSpace(opts.RunInput.IconURL) != "" ||
		strings.TrimSpace(opts.RunInput.Browser) != "" ||
		strings.TrimSpace(opts.RunInput.URLTemplate) != "" ||
		strings.TrimSpace(opts.RunInput.ExtraFlags) != "" ||
		strings.TrimSpace(opts.RunInput.StartupWMClass) != ""
}

func parseConfigCommand(args []string) (RuntimeOptions, error) {
	fs := flag.NewFlagSet("desktopify-lite config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var opts RuntimeOptions
	opts.Command = CommandConfig
	var showHelp bool
	var showVersion bool
	var defaultBrowser trackedString
	var defaultURLTemplate trackedString
	var defaultExtraFlags trackedString
	var defaultProxy trackedString
	var disableGoogleFavicon trackedBool
	var withDebug trackedBool

	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.BoolVar(&showVersion, "v", false, "show version")
	fs.Var(&defaultBrowser, "default_browser", "default browser binary")
	fs.Var(&defaultURLTemplate, "default_url_template", "default URL template")
	fs.Var(&defaultExtraFlags, "default_extra_flags", "default extra browser flags")
	fs.Var(&defaultProxy, "default_proxy", "default proxy URL")
	fs.Var(&disableGoogleFavicon, "disable_google_favicon", "disable Google favicon lookup")
	fs.Var(&withDebug, "with_debug", "enable or disable debug output")

	if err := fs.Parse(args); err != nil {
		return RuntimeOptions{}, err
	}
	if showHelp {
		return RuntimeOptions{Command: CommandHelp}, nil
	}
	if showVersion {
		return RuntimeOptions{Command: CommandVersion}, nil
	}
	if fs.NArg() != 0 {
		return RuntimeOptions{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	if defaultBrowser.set {
		opts.ConfigUpdates.DefaultBrowser = &defaultBrowser.value
	}
	if defaultURLTemplate.set {
		opts.ConfigUpdates.DefaultURLTemplate = &defaultURLTemplate.value
	}
	if defaultExtraFlags.set {
		opts.ConfigUpdates.DefaultExtraFlags = &defaultExtraFlags.value
	}
	if defaultProxy.set {
		opts.ConfigUpdates.DefaultProxy = &defaultProxy.value
	}
	if disableGoogleFavicon.set {
		opts.ConfigUpdates.DisableGoogleFavicon = &disableGoogleFavicon.value
	}
	if withDebug.set {
		opts.ConfigUpdates.WithDebug = &withDebug.value
	}
	if opts.ConfigUpdates.DefaultProxy != nil {
		if err := ValidateProxyURL(*opts.ConfigUpdates.DefaultProxy); err != nil {
			return RuntimeOptions{}, err
		}
	}

	return opts, nil
}

func parseConfigResetCommand(args []string) (RuntimeOptions, error) {
	fs := flag.NewFlagSet("desktopify-lite config-reset", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var showHelp bool
	var showVersion bool
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.BoolVar(&showVersion, "v", false, "show version")

	if err := fs.Parse(args); err != nil {
		return RuntimeOptions{}, err
	}
	if showHelp {
		return RuntimeOptions{Command: CommandHelp}, nil
	}
	if showVersion {
		return RuntimeOptions{Command: CommandVersion}, nil
	}
	if fs.NArg() != 0 {
		return RuntimeOptions{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	return RuntimeOptions{Command: CommandConfigReset}, nil
}

func EffectiveProxy(opts RuntimeOptions, cfg Config) string {
	if strings.TrimSpace(opts.Proxy) != "" {
		return strings.TrimSpace(opts.Proxy)
	}
	return strings.TrimSpace(cfg.DefaultProxy)
}

func ValidateProxyURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid proxy url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("invalid proxy url: must include scheme and host")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return nil
	default:
		return fmt.Errorf("invalid proxy url: unsupported scheme %q", parsed.Scheme)
	}
}
