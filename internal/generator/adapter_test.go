package generator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapterGenerator_isErrorDefined(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		errorName string
		want      bool
		before    func(*AdapterGenerator)
	}{
		{
			name:      "error defined in var block",
			filePath:  writeTestFile(t, "package foo\nvar ErrFoo = errors.New(\"foo\")\n"),
			errorName: "ErrFoo",
			want:      true,
		},
		{
			name:      "error not defined",
			filePath:  writeTestFile(t, "package foo\nvar ErrFoo = errors.New(\"foo\")\n"),
			errorName: "ErrBar",
			want:      false,
		},
		{
			name:      "file does not exist",
			filePath:  "/nonexistent/path/foo.go",
			errorName: "ErrFoo",
			want:      false,
		},
		{
			name:      "package load error on directory path",
			filePath:  t.TempDir(),
			errorName: "ErrFoo",
			want:      false,
		},
		{
			name:      "const declaration ignored",
			filePath:  writeTestFile(t, "package foo\nconst ErrFoo = \"error\"\n"),
			errorName: "ErrFoo",
			want:      false,
		},
		{
			name:      "empty file",
			filePath:  writeTestFile(t, "package foo\n"),
			errorName: "ErrFoo",
			want:      false,
		},
		{
			name:      "multiple vars one matches",
			filePath:  writeTestFile(t, "package foo\nvar (\n\tErrOne = errors.New(\"one\")\n\tErrTwo = errors.New(\"two\")\n)\n"),
			errorName: "ErrTwo",
			want:      true,
		},
		{
			name:      "var block but error absent",
			filePath:  writeTestFile(t, "package foo\nvar (\n\tErrOne = errors.New(\"one\")\n\tErrTwo = errors.New(\"two\")\n)\n"),
			errorName: "ErrThree",
			want:      false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewAdapterGenerator(nil)
			if tt.before != nil {
				tt.before(s)
			}
			r := s.isErrorDefined(tt.filePath, tt.errorName)
			assert.Equal(t, tt.want, r)
		})
	}
}

func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := dir + "/test.go"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
