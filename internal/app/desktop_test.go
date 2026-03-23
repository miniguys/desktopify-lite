package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestSanitizeFileStem(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"My App", "My_App"},
		{"__ok__", "__ok__"},
		{"Café", "Café"},
		{"!!!", "webapp"},
		{"", "webapp"},
	}

	for _, tt := range tests {
		got := sanitizeFileStem(tt.in)
		if got != tt.want {
			t.Fatalf("sanitizeFileStem(%q)=%q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBuildExec_UsesConfigDefaults(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "chromium",
		DefaultURLTemplate: "--app={url}",
		DefaultExtraFlags:  "--profile-directory=Default",
	}

	in := Input{URL: "https://example.com"}

	args, err := BuildExecArgs(in, cfg)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"chromium",
		"--profile-directory=Default",
		"--app=https://example.com",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildExec_OverridesConfig(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "chromium",
		DefaultURLTemplate: "--app={url}",
	}

	in := Input{
		URL:         "https://example.com",
		Browser:     "firefox",
		URLTemplate: "{url}",
		ExtraFlags:  "--private-window",
	}

	args, err := BuildExecArgs(in, cfg)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"firefox",
		"--private-window",
		"https://example.com",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildExec_TemplateMustContainURL(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "chromium",
		DefaultURLTemplate: "invalid",
	}

	in := Input{URL: "https://example.com"}

	_, err := BuildExecArgs(in, cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildExec_URLTemplateCanExpandToMultipleArgs(t *testing.T) {
	cfg := Config{DefaultBrowser: "chromium", DefaultURLTemplate: "{url}"}
	in := Input{URL: "https://example.com", URLTemplate: `--new-window --app="{url}"`}

	args, err := BuildExecArgs(in, cfg)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"chromium", "--new-window", "--app=https://example.com"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected args: got %#v want %#v", args, want)
	}
}

func TestBuildExecEscapesPercentAndReservedChars(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "firefox",
		DefaultURLTemplate: "{url}",
	}

	in := Input{URL: "https://example.com/a%20b?a=1&b=2"}

	exec, err := BuildExec(in, cfg)
	if err != nil {
		t.Fatal(err)
	}

	want := `firefox "https://example.com/a%%20b?a=1&b=2"`
	if exec != want {
		t.Fatalf("unexpected exec: %q", exec)
	}
}

func TestEscapeExecArg_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "--app=https://example.com", want: "--app=https://example.com"},
		{name: "space", in: "Profile 1", want: `"Profile 1"`},
		{name: "percent", in: "https://example.com/a%20b", want: `"https://example.com/a%%20b"`},
		{name: "backslashes", in: `C:\path\to\app`, want: `"C:\\path\\to\\app"`},
		{name: "quotes and dollar", in: `say "hi" $HOME`, want: `"say \"hi\" $$HOME"`},
		{name: "backtick", in: "`cmd`", want: "\"\\`cmd\\`\""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeExecArg(tc.in); got != tc.want {
				t.Fatalf("escapeExecArg(%q)=%q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildDesktopEntryOmitsVersionAndSupportsWMClass(t *testing.T) {
	entry, _, err := BuildDesktopEntry(Input{
		URL:            "https://example.com",
		Name:           "Example",
		StartupWMClass: "ExampleApp",
	}, "", Config{DefaultBrowser: "chromium", DefaultURLTemplate: "{url}"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(entry, "Version=") {
		t.Fatalf("unexpected Version field: %q", entry)
	}
	if !strings.Contains(entry, "StartupWMClass=ExampleApp") {
		t.Fatalf("missing StartupWMClass: %q", entry)
	}
	if strings.Contains(entry, "\nIcon=") {
		t.Fatalf("icon should be omitted when empty: %q", entry)
	}
}

func TestSplitCommandLineErrors(t *testing.T) {
	if _, err := splitCommandLine(`firefox "abc`); err == nil {
		t.Fatalf("expected unclosed quote error")
	}
	if _, err := splitCommandLine(`firefox abc\`); err == nil {
		t.Fatalf("expected dangling escape error")
	}
}

func TestBuildExec_ExtraFlagsWithQuotes(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "chromium",
		DefaultURLTemplate: "{url}",
	}

	in := Input{
		URL:        "https://example.com",
		ExtraFlags: `--profile-directory="Profile 1"`,
	}

	args, err := BuildExecArgs(in, cfg)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"chromium",
		"--profile-directory=Profile 1",
		"https://example.com",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestLauncherFileStemDisambiguatesLossyNames(t *testing.T) {
	plain := launcherFileStem("Foo")
	bang := launcherFileStem("Foo!")
	question := launcherFileStem("Foo?")

	if plain != "Foo" {
		t.Fatalf("plain stem=%q, want %q", plain, "Foo")
	}
	if bang == plain {
		t.Fatalf("expected Foo! stem to differ from plain stem, got %q", bang)
	}
	if bang == question {
		t.Fatalf("expected distinct stems for lossy names, got %q", bang)
	}
	if !strings.HasPrefix(bang, "Foo-") {
		t.Fatalf("expected hashed stem prefix for Foo!, got %q", bang)
	}
}

func TestLauncherFileStemKeepsCommonSpaceOnlyNamesStable(t *testing.T) {
	if got := launcherFileStem("Example App"); got != "Example_App" {
		t.Fatalf("launcherFileStem=%q, want %q", got, "Example_App")
	}
}
