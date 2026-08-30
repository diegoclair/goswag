// Command goswag orchestrates the full swagger-generation pipeline so
// users don't need a hand-rolled Makefile. It is a thin wrapper around:
//
//  1. `go run <input>/main.go`     — runs the user's stub generator (which
//     calls goswag.GenerateSwagger() internally)
//  2. `swag init ...`              — generates the OpenAPI JSON/YAML from
//     the annotated stub
//  3. `swag fmt -d <input>/` + gofmt — formats the annotations in place
//
// If `swag` is not on PATH, the CLI installs it automatically (it's a hard
// dependency anyway). All paths default to the convention documented in
// the README, so `goswag docs` with no flags works for the recommended
// project layout.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	swagInstallPath = "github.com/swaggo/swag/cmd/swag@latest"

	// pdlAuto is the sentinel for "detect from the generated goswag.go imports".
	// Negative because the swag --pdl range is 0..3.
	pdlAuto = -1

	// defaultOutputTypes mirrors swag's own default so the pipeline is unchanged
	// for anyone who never touches the flag.
	defaultOutputTypes = "go,json,yaml"
)

var supportedOutputTypes = map[string]bool{"go": true, "json": true, "yaml": true, "yml": true}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "docs":
		if err := runDocs(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "goswag: "+err.Error())
			os.Exit(1)
		}
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "goswag: unknown command %q\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

type docsConfig struct {
	input         string
	output        string
	outputTypes   string
	pdl           int
	parseInternal bool
	skipFormat    bool
}

func newDocsFlagSet(cfg *docsConfig) *flag.FlagSet {
	fs := flag.NewFlagSet("docs", flag.ContinueOnError)
	fs.StringVar(&cfg.input, "input", "./goswag", "directory containing the main.go that calls GenerateSwagger()")
	fs.StringVar(&cfg.input, "i", "./goswag", "shorthand for --input")
	fs.StringVar(&cfg.output, "output", "./docs", "directory where swag will write the OpenAPI files")
	fs.StringVar(&cfg.output, "o", "./docs", "shorthand for --output")
	fs.StringVar(&cfg.outputTypes, "output-types", defaultOutputTypes, "comma-separated swag --ot types (go, json, yaml, yml)")
	fs.IntVar(&cfg.pdl, "pdl", pdlAuto, "swag --pdl (0..3); default auto-detects from imports in the generated stub")
	fs.BoolVar(&cfg.parseInternal, "parse-internal", true, "pass --parseInternal to swag init")
	fs.BoolVar(&cfg.skipFormat, "skip-format", false, "skip the `swag fmt` + gofmt step at the end")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: goswag docs [flags]")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Runs the full swagger generation pipeline:")
		fmt.Fprintln(fs.Output(), "  1. go run <input>/main.go    (generates the annotated stub)")
		fmt.Fprintln(fs.Output(), "  2. swag init                 (generates the OpenAPI spec)")
		fmt.Fprintln(fs.Output(), "  3. swag fmt + gofmt          (formats annotations in place)")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Flags:")
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Auto-detected --pdl:")
		fmt.Fprintln(fs.Output(), "  When --pdl is not set, goswag inspects the imports of the generated")
		fmt.Fprintln(fs.Output(), "  stub. If any import points outside the user's module (e.g. types from")
		fmt.Fprintln(fs.Output(), "  diegoclair/go_utils referenced in @Failure annotations), --pdl=1 is")
		fmt.Fprintln(fs.Output(), "  used so swag can resolve those struct definitions. Otherwise it stays")
		fmt.Fprintln(fs.Output(), "  at 0. Pass --pdl=N explicitly to override.")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Output types:")
		fmt.Fprintln(fs.Output(), "  --output-types controls which files swag writes. Dropping \"go\" skips the")
		fmt.Fprintln(fs.Output(), "  generated docs.go, so a project that only serves swagger.json/yaml does not")
		fmt.Fprintln(fs.Output(), "  need github.com/swaggo/swag as a direct dependency.")
	}

	return fs
}

func runDocs(args []string) error {
	cfg := docsConfig{}
	fs := newDocsFlagSet(&cfg)

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError returns ErrHelp on -h/--help; that's not a failure.
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	normalized, err := normalizeOutputTypes(cfg.outputTypes)
	if err != nil {
		return err
	}
	cfg.outputTypes = normalized

	mainFile := filepath.Join(cfg.input, "main.go")
	if _, err := os.Stat(mainFile); err != nil {
		return fmt.Errorf("input main.go not found at %s — pass --input to point at the right directory", mainFile)
	}

	if err := ensureSwag(); err != nil {
		return err
	}

	fmt.Printf("=====> goswag: generating stub (go run %s)\n", mainFile)
	if err := run(cfg.input, "go", "run", "main.go"); err != nil {
		return fmt.Errorf("go run failed: %w", err)
	}

	pdl := cfg.pdl
	autodetected := false
	if pdl == pdlAuto {
		detected, reason, err := detectPDL(cfg.input)
		if err != nil {
			// Detection failure is not fatal — fall back to the conservative
			// "include external models" level which works for almost every
			// realistic project.
			fmt.Fprintf(os.Stderr, "goswag: pdl autodetect failed (%v), falling back to --pdl=1\n", err)
			pdl = 1
		} else {
			pdl = detected
			fmt.Printf("=====> goswag: auto-detected --pdl=%d (%s)\n", pdl, reason)
		}
		autodetected = true
	}

	swagArgs := buildSwagArgs(cfg, pdl, mainFile)
	fmt.Printf("=====> goswag: running swag init -> %s\n", cfg.output)
	if err := run("", "swag", swagArgs...); err != nil {
		// swag init's error output is the typical signal a user gets that
		// they need a higher --pdl (e.g. a dep with @Router annotations on
		// imported handlers). The autodetect only chooses 0 or 1, so if the
		// run came from autodetect and failed, point the user at the
		// override before bailing.
		if autodetected {
			fmt.Fprintln(os.Stderr,
				"goswag: swag init failed under auto-detected --pdl. If swag reported missing types or unresolved routes, try `goswag docs --pdl=2` or `--pdl=3` to make swag parse operations / all of your external deps.")
		}
		return fmt.Errorf("swag init failed: %w", err)
	}

	if !cfg.skipFormat {
		fmt.Printf("=====> goswag: running swag fmt on %s\n", cfg.input)
		if err := run("", "swag", "fmt", "-d", cfg.input); err != nil {
			return fmt.Errorf("swag fmt failed: %w", err)
		}
		// swag fmt indents annotations with a tab after "//", which gofmt rewrites
		// to a space; normalising here keeps the file stable under editor save.
		fmt.Printf("=====> goswag: running gofmt on %s\n", cfg.input)
		if err := run("", "gofmt", "-w", cfg.input); err != nil {
			return fmt.Errorf("gofmt failed: %w", err)
		}
	}

	fmt.Println("=====> goswag: done")
	return nil
}

func buildSwagArgs(cfg docsConfig, pdl int, mainFile string) []string {
	args := []string{"init", "--pdl", strconv.Itoa(pdl), "-g", mainFile, "-o", cfg.output, "--ot", cfg.outputTypes}
	if cfg.parseInternal {
		args = append(args, "--parseInternal")
	}
	return args
}

// normalizeOutputTypes rejects here what swag only logs and then ignores, which
// would otherwise leave the user with a silently missing file and a zero exit code.
func normalizeOutputTypes(value string) (string, error) {
	parts := strings.Split(value, ",")
	types := make([]string, 0, len(parts))
	for _, part := range parts {
		t := strings.ToLower(strings.TrimSpace(part))
		if !supportedOutputTypes[t] {
			return "", fmt.Errorf("invalid --output-types value %q: supported types are go, json, yaml, yml", strings.TrimSpace(part))
		}
		types = append(types, t)
	}
	return strings.Join(types, ","), nil
}

// detectPDL inspects the generated goswag.go to decide which --pdl level
// swag needs. It returns the chosen level plus a short human-readable
// reason for the log.
//
// Rationale: swag won't enter GOMODCACHE without --pdl >= 1, so any type
// referenced in an annotation that lives in an external dependency (e.g.
// `@Failure 400 {object} resterrors.restErr`) makes generation fail with
// "cannot find type definition". We detect this by parsing the imports of
// the generated stub — if any non-stdlib import points outside the user's
// module, we need --pdl=1.
//
// We deliberately only choose between 0 and 1. Levels 2 and 3 are only
// useful when the user imports a dependency that *defines @Router routes*
// (extremely rare) — those cases stay as a manual override via --pdl=N.
func detectPDL(input string) (int, string, error) {
	goswagFile := filepath.Join(input, "goswag.go")
	if _, err := os.Stat(goswagFile); err != nil {
		return 0, "", fmt.Errorf("generated stub not found at %s", goswagFile)
	}

	modulePath, err := readModulePath()
	if err != nil {
		return 0, "", err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, goswagFile, nil, parser.ImportsOnly)
	if err != nil {
		return 0, "", fmt.Errorf("parsing %s: %w", goswagFile, err)
	}

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if isStdlibImport(path) {
			continue
		}
		if path == modulePath || strings.HasPrefix(path, modulePath+"/") {
			continue
		}
		return 1, "external dep " + path + " referenced", nil
	}
	return 0, "no external deps referenced in stub", nil
}

// isStdlibImport returns true for standard-library import paths. Go's
// stdlib uses single-segment-or-dotless first segments (e.g. "strings",
// "net/http", "go/parser"); third-party imports always include a domain
// (e.g. "github.com/...", "golang.org/x/..."), so the presence of a "." in
// the first path segment is a reliable discriminator.
func isStdlibImport(path string) bool {
	first := path
	if before, _, ok := strings.Cut(path, "/"); ok {
		first = before
	}
	return !strings.Contains(first, ".")
}

// readModulePath reads the module declaration from the go.mod nearest the
// current working directory, walking up the filesystem if necessary.
// Implemented without golang.org/x/mod to keep the CLI's dependency
// footprint minimal — the module line we care about is the first line
// matching `module <path>` and that grammar has been stable for years.
func readModulePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; {
		gomod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(gomod); err == nil {
			for line := range strings.SplitSeq(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
				}
			}
			return "", fmt.Errorf("%s has no module declaration", gomod)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// ensureSwag guarantees that the `swag` binary is on PATH, installing it
// if missing. Installing a hard dependency without prompting is acceptable
// here because the CLI exists specifically to orchestrate swag.
func ensureSwag() error {
	if _, err := exec.LookPath("swag"); err == nil {
		return nil
	}
	fmt.Println("=====> goswag: swag binary not found; installing", swagInstallPath)
	if err := run("", "go", "install", swagInstallPath); err != nil {
		return fmt.Errorf("failed to install swag: %w (install it manually with: go install %s)", err, swagInstallPath)
	}
	if _, err := exec.LookPath("swag"); err != nil {
		return fmt.Errorf("swag still not on PATH after install — ensure $(go env GOPATH)/bin is in your PATH")
	}
	return nil
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func printVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("goswag: version information unavailable")
		return
	}
	fmt.Printf("goswag %s\n", info.Main.Version)
}

func printUsage() {
	fmt.Println(`goswag — swagger generation pipeline for Go APIs

Usage:
  goswag <command> [flags]

Commands:
  docs       Run the full swagger pipeline (go run + swag init + swag fmt)
  version    Print the installed CLI version
  help       Show this message

Run "goswag docs --help" for command-specific flags.

Updating:
  CLI:  go install github.com/diegoclair/goswag/v2/cmd/goswag@latest
  Lib:  go get -u github.com/diegoclair/goswag/v2 && go mod tidy`)
}
