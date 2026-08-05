package version

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SplashCheckFn func(*testing.T, string, string)

var checkSplash = func(fns ...SplashCheckFn) []SplashCheckFn { return fns }

type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestSplash(t *testing.T) {

	checkContains := func(want string) SplashCheckFn {
		return func(t *testing.T, std, _ string) {
			t.Helper()
			assert.Containsf(t, std, want, "%s expected in stdout", want)
		}
	}

	checkNotContains := func(want string) SplashCheckFn {
		return func(t *testing.T, std, _ string) {
			t.Helper()
			assert.NotContainsf(t, std, want, "%s not expected in stdout", want)
		}
	}

	checkError := func(want string) SplashCheckFn {
		return func(t *testing.T, _, err string) {
			t.Helper()
			if want != "" {
				assert.Containsf(t, err, want, "%s expected in error", want)
				return
			}
			assert.Emptyf(t, err, "error not expected: %s", err)
		}
	}

	tests := []struct {
		name    string
		version *VersionInfo
		buff    io.Writer
		checks  []SplashCheckFn
	}{
		{
			name: "success with extra",
			buff: &bytes.Buffer{},
			version: &VersionInfo{
				Version:   "v1.2.3-rc1",
				Major:     1,
				Minor:     2,
				Patch:     3,
				Extra:     "rc1",
				Commit:    "abc123def",
				BuildDate: "2025-01-15T10:00:00Z",
			},
			checks: checkSplash(
				checkError(""),
				checkContains("Version: 1.2.3-rc1"),
				checkContains("Build: 2025-01-15T10:00:00Z"),
				checkContains("Commit: abc123def"),
			),
		},
		{
			name: "success without extra",
			buff: &bytes.Buffer{},
			version: &VersionInfo{
				Version:   "v1.2.3",
				Major:     1,
				Minor:     2,
				Patch:     3,
				Extra:     "",
				Commit:    "deadbeef",
				BuildDate: "2025-06-01T00:00:00Z",
			},
			checks: checkSplash(
				checkError(""),
				checkContains("Version: 1.2.3"),
				checkNotContains("Version: 1.2.3-"),
				checkContains("Build: 2025-06-01T00:00:00Z"),
				checkContains("Commit: deadbeef"),
			),
		},
		{
			name: "success with default version",
			buff: &bytes.Buffer{},
			checks: checkSplash(
				checkError(""),
				checkContains("Version: 0.0.0"),
				checkContains("Build: unknown"),
				checkContains("Commit: unknown"),
			),
		},
		{
			name: "execute error writes to errWriter",
			buff: &errWriter{err: assert.AnError},
			version: &VersionInfo{
				Version:   "v0.0.1",
				Major:     0,
				Minor:     0,
				Patch:     1,
				BuildDate: "unknown",
				Commit:    "unknown",
			},
			checks: checkSplash(
				checkError("Error executing template"),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.version != nil {
				defer func(orig *VersionInfo) { v = orig }(v)
				v = tt.version
			}

			errBuf := &bytes.Buffer{}

			Splash(tt.buff, errBuf)

			var stdStr string
			if buf, ok := tt.buff.(*bytes.Buffer); ok {
				stdStr = buf.String()
			}

			for _, c := range tt.checks {
				c(t, stdStr, errBuf.String())
			}
		})
	}
}
