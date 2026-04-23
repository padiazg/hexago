package testgen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	Binary     = "go-testgen"
	MinVersion = "v0.1.0"
)

// FuncSummary mirrors go-testgen's analyzer.FuncSummary (JSON subset we need).
type FuncSummary struct {
	FuncSpec       string         `json:"funcSpec"`
	TestExists     bool           `json:"testExists"`
	SuggestedStyle string         `json:"suggestedStyle,omitempty"`
	InterfaceDeps  []InterfaceDep `json:"interfaceDeps"`
}

// InterfaceDep mirrors go-testgen's analyzer.InterfaceDep.
type InterfaceDep struct {
	MockFrom   string `json:"mockFrom"`
	MockExists bool   `json:"mockExists"`
}

// reportResult mirrors go-testgen's analyzer.ScanResult (fields we need).
type reportResult struct {
	Funcs []FuncSummary `json:"funcs"`
}

// Available checks whether go-testgen is on PATH and meets MinVersion.
// Returns the version string, whether it's usable, and any error.
func Available(ctx context.Context) (version string, ok bool, err error) {
	cmd := exec.CommandContext(ctx, Binary, "version", "--simple")
	out, execErr := cmd.Output()
	if execErr != nil {
		return "", false, execErr
	}
	version = strings.TrimSpace(string(out))
	// go-testgen emits versions without a leading "v"; semver.IsValid requires one.
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !semver.IsValid(version) {
		return version, false, fmt.Errorf("unexpected version string: %q", version)
	}
	if semver.Compare(version, MinVersion) < 0 {
		return version, false, fmt.Errorf("%s is below minimum required %s", version, MinVersion)
	}
	return version, true, nil
}

// Report runs `go-testgen report <pkgPattern> --format json` in projectRoot
// and returns the decoded FuncSummary slice.
func Report(ctx context.Context, projectRoot, pkgPattern string) ([]FuncSummary, error) {
	cmd := exec.CommandContext(ctx, Binary, "report", pkgPattern, "--format", "json")
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	var result reportResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("report: decode json: %w", err)
	}
	return result.Funcs, nil
}

// Gen runs `go-testgen gen` for a single function using the args derived
// from its FuncSummary (pkg pattern, funcSpec, style, mock-from flags).
func Gen(ctx context.Context, projectRoot, pkgPattern string, fn FuncSummary) error {
	args := []string{"gen", pkgPattern, fn.FuncSpec}
	if fn.SuggestedStyle != "" && fn.SuggestedStyle != "check" {
		args = append(args, "--test-style", fn.SuggestedStyle)
	}
	for _, dep := range fn.InterfaceDeps {
		if !dep.MockExists {
			args = append(args, "--mock-from", dep.MockFrom)
		}
	}
	cmd := exec.CommandContext(ctx, Binary, args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GenerateForPackage orchestrates test generation for all untested exported
// functions in pkgPattern: checks availability, runs report, then gen per entry.
// On any failure it warns to stderr and returns nil — never blocks the caller.
func GenerateForPackage(ctx context.Context, projectRoot, pkgPattern string) error {
	_, ok, err := Available(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  go-testgen not available: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Install: go install github.com/padiazg/go-testgen@latest\n")
		return nil
	}
	if !ok {
		return nil
	}

	funcs, err := Report(ctx, projectRoot, pkgPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  go-testgen report failed: %v\n", err)
		return nil
	}

	for _, fn := range funcs {
		if fn.TestExists {
			continue
		}
		fmt.Printf("🧪 Generating test: %s\n", fn.FuncSpec)
		if err := Gen(ctx, projectRoot, pkgPattern, fn); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  go-testgen gen %s: %v\n", fn.FuncSpec, err)
		}
	}

	return nil
}
