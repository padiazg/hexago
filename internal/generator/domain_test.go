package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type parsePathCheckFn func(*testing.T, map[string]string)

var checkparsePath = func(fns ...parsePathCheckFn) []parsePathCheckFn { return fns }

func Test_parsePath(t *testing.T) {
	tests := []struct {
		name    string
		module  string
		content string
		checks  []parsePathCheckFn
	}{
		{
			name:    "exports a struct type",
			module:  "github.com/padiazg/hexago",
			content: "package user\ntype User struct {}",
			checks: checkparsePath(
				assertTypeExists("User"),
			),
		},
		{
			name:    "exports multiple types",
			module:  "github.com/myorg/myservice",
			content: "package money\ntype Money struct{}\ntype Currency struct{}",
			checks: checkparsePath(
				assertTypeExists("Money"),
				assertTypeExists("Currency"),
			),
		},
		{
			name:    "skips unexported types",
			module:  "github.com/padiazg/hexago",
			content: "package user\ntype User struct {}\ntype internalType struct {}",
			checks: checkparsePath(
				assertTypeExists("User"),
				assertTypeNotExists("internalType"),
			),
		},
		{
			name:    "skips non-type declarations",
			module:  "github.com/padiazg/hexago",
			content: "package user\nimport \"fmt\"\ntype User struct{}\nfunc NewUser() *User { return nil }",
			checks: checkparsePath(
				assertTypeExists("User"),
				assertTypeNotExists("NewUser"),
			),
		},
		{
			name:    "handles file at domain root",
			module:  "github.com/padiazg/hexago",
			content: "package domain\ntype Status struct{}",
			checks: checkparsePath(
				assertTypeExists("Status"),
			),
		},
		{
			name:    "handles missing file gracefully",
			module:  "github.com/padiazg/hexago",
			content: "",
			checks: checkparsePath(
				assertTypeNotExists("Anything"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != "" {
				tmpDir := t.TempDir()
				err := os.WriteFile(filepath.Join(tmpDir, "user.go"), []byte(tt.content), 0o644)
				if err != nil {
					t.Fatalf("setup: %v", err)
				}
				path = filepath.Join(tmpDir, "user.go")
			}

			index := map[string]string{}
			modulePath := tt.module + "/internal/core/domain/user"
			parsePath(path, tt.module, modulePath, func(i, importPath string) { index[i] = importPath })
			for _, c := range tt.checks {
				c(t, index)
			}
		})
	}
}

func assertTypeExists(key string) parsePathCheckFn {
	return func(t *testing.T, index map[string]string) {
		t.Helper()
		_, ok := index[key]
		assert.True(t, ok, "index should contain key %q", key)
	}
}

func assertTypeNotExists(key string) parsePathCheckFn {
	return func(t *testing.T, index map[string]string) {
		t.Helper()
		_, ok := index[key]
		assert.False(t, ok, "index should not contain key %q", key)
	}
}

func assertIndexEmpty() resolveDomainImportsFn {
	return func(t *testing.T, index map[string]string) {
		t.Helper()
		assert.Empty(t, index, "index should be empty")
	}
}

func assertIndexEntry(key, val string) resolveDomainImportsFn {
	return func(t *testing.T, index map[string]string) {
		t.Helper()
		got, ok := index[key]
		assert.True(t, ok, "index should contain key %q", key)
		assert.Equal(t, val, got)
	}
}

func assertIndexNotContains(key string) resolveDomainImportsFn {
	return func(t *testing.T, index map[string]string) {
		t.Helper()
		_, ok := index[key]
		assert.False(t, ok, "index should not contain key %q", key)
	}
}

type resolveDomainImportsFn func(*testing.T, map[string]string)

var checkresolveDomainImports = func(fns ...resolveDomainImportsFn) []resolveDomainImportsFn { return fns }

func Test_resolveDomainImports(t *testing.T) {
	tests := []struct {
		name     string
		module   string
		entities map[string]string
		checks   []resolveDomainImportsFn
	}{
		{
			name:     "empty domain directory",
			module:   "github.com/padiazg/hexago",
			entities: map[string]string{},
			checks: checkresolveDomainImports(
				assertIndexEmpty(),
			),
		},
		{
			name:   "single entity",
			module: "github.com/padiazg/hexago",
			entities: map[string]string{
				"user/user.go": "package user\ntype User struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexEntry("User", "github.com/padiazg/hexago/internal/core/domain/user"),
			),
		},
		{
			name:   "multiple entities in different packages",
			module: "github.com/myorg/myservice",
			entities: map[string]string{
				"user/user.go":       "package user\ntype User struct {}",
				"money/money.go":     "package money\ntype Money struct {}",
				"address/address.go": "package address\ntype Address struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexEntry("User", "github.com/myorg/myservice/internal/core/domain/user"),
				assertIndexEntry("Money", "github.com/myorg/myservice/internal/core/domain/money"),
				assertIndexEntry("Address", "github.com/myorg/myservice/internal/core/domain/address"),
			),
		},
		{
			name:   "entity at domain root",
			module: "github.com/padiazg/hexago",
			entities: map[string]string{
				"user.go": "package domain\ntype Status struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexEntry("Status", "github.com/padiazg/hexago/internal/core/domain/"),
			),
		},
		{
			name:   "skips unexported types",
			module: "github.com/padiazg/hexago",
			entities: map[string]string{
				"user/user.go": "package user\ntype User struct {}\ntype internalUser struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexEntry("User", "github.com/padiazg/hexago/internal/core/domain/user"),
				assertIndexNotContains("internalUser"),
			),
		},
		{
			name:   "skips test files",
			module: "github.com/padiazg/hexago",
			entities: map[string]string{
				"user/user_test.go": "package user\ntype TestType struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexNotContains("TestType"),
			),
		},
		{
			name:   "multiple types in single package",
			module: "example.com/project",
			entities: map[string]string{
				"order/order.go": "package order\ntype Order struct {}\ntype OrderItem struct {}",
			},
			checks: checkresolveDomainImports(
				assertIndexEntry("Order", "example.com/project/internal/core/domain/order"),
				assertIndexEntry("OrderItem", "example.com/project/internal/core/domain/order"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			domainDir := filepath.Join(tmpDir, "internal", "core", "domain")
			for relPath, content := range tt.entities {
				fpath := filepath.Join(domainDir, relPath)
				if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
					t.Fatalf("create dir: %v", err)
				}
				if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
					t.Fatalf("write file: %v", err)
				}
			}

			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("get cwd: %v", err)
			}
			defer os.Chdir(cwd)
			os.Chdir(tmpDir)

			r := resolveDomainImports(tt.module, tmpDir)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

type collectImportsFn func(*testing.T, map[string]bool)

var checkcollectImports = func(fns ...collectImportsFn) []collectImportsFn { return fns }

func Test_collectImports(t *testing.T) {
	contains := func(want string) collectImportsFn {
		return func(t *testing.T, m map[string]bool) {
			t.Helper()
			assert.Contains(t, m, want)
		}
	}

	notContains := func(want string) collectImportsFn {
		return func(t *testing.T, m map[string]bool) {
			t.Helper()
			assert.NotContains(t, m, want)
		}
	}

	module := "github.com/padiazg/hexago"

	tests := []struct {
		name        string
		selfPkg     string
		baseImports []string
		fields      []Field
		index       map[string]string
		checks      []collectImportsFn
	}{
		{
			name:        "empty fields returns only base imports",
			selfPkg:     module + "/internal/core/domain/user",
			baseImports: []string{"fmt", "time"},
			fields:      nil,
			index:       map[string]string{},
			checks: checkcollectImports(
				contains("fmt"),
				contains("time"),
			),
		},
		{
			name:        "base imports always included",
			selfPkg:     module + "/internal/core/domain/user",
			baseImports: []string{"context", "net/http"},
			fields:      nil,
			checks: checkcollectImports(
				contains("context"),
				contains("net/http"),
			),
		},
		{
			name:    "builtin types add no imports",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "ID", Type: "string"},
				{Name: "Age", Type: "int"},
				{Name: "Active", Type: "bool"},
				{Name: "Data", Type: "[]byte"},
				{Name: "Rune", Type: "rune"},
				{Name: "Err", Type: "error"},
			},
			checks: checkcollectImports(),
		},
		{
			name:    "time.Time adds time import",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "CreatedAt", Type: "time.Time"},
			},
			checks: checkcollectImports(
				contains("time"),
			),
		},
		{
			name:    "uuid.UUID adds uuid import",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "ID", Type: "uuid.UUID"},
			},
			checks: checkcollectImports(
				contains("github.com/google/uuid"),
			),
		},
		{
			name:    "json.RawMessage adds json import",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "Data", Type: "json.RawMessage"},
			},
			checks: checkcollectImports(
				contains("encoding/json"),
			),
		},
		{
			name:    "json.Number adds json import",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "Amount", Type: "json.Number"},
			},
			checks: checkcollectImports(
				contains("encoding/json"),
			),
		},
		{
			name:    "custom domain type from index",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "Email", Type: "email.Email"},
			},
			index: map[string]string{
				"email.Email": module + "/internal/core/domain/email",
			},
			checks: checkcollectImports(
				contains("github.com/padiazg/hexago/internal/core/domain/email"),
			),
		},
		{
			name:    "skips self-pkg types",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "Status", Type: "Status"},
			},
			index: map[string]string{
				"Status": module + "/internal/core/domain/user",
			},
			checks: checkcollectImports(
				notContains("github.com/padiazg/hexago/internal/core/domain/user"),
			),
		},
		{
			name:    "mixed builtin and domain types",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "ID", Type: "string"},
				{Name: "Email", Type: "email.Email"},
				{Name: "CreatedAt", Type: "time.Time"},
				{Name: "Token", Type: "uuid.UUID"},
				{Name: "Raw", Type: "json.RawMessage"},
			},
			index: map[string]string{
				"email.Email": module + "/internal/core/domain/email",
			},
			checks: checkcollectImports(
				contains("time"),
				contains("github.com/google/uuid"),
				contains("encoding/json"),
				contains("github.com/padiazg/hexago/internal/core/domain/email"),
			),
		},
		{
			name:        "base imports not duplicated by fields",
			selfPkg:     module + "/internal/core/domain/user",
			baseImports: []string{"fmt"},
			fields: []Field{
				{Name: "Name", Type: "string"},
			},
			checks: checkcollectImports(
				contains("fmt"),
			),
		},
		{
			name:    "unexported type field adds nothing",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "inner", Type: "innerType"},
			},
			checks: checkcollectImports(
				notContains("innerType"),
			),
		},
		{
			name:    "slice of custom type",
			selfPkg: module + "/internal/core/domain/user",
			fields: []Field{
				{Name: "Tags", Type: "[]tag.Tag"},
			},
			index: map[string]string{
				"tag.Tag": module + "/internal/core/domain/tag",
			},
			checks: checkcollectImports(
				contains("github.com/padiazg/hexago/internal/core/domain/tag"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := collectImports(tt.selfPkg, tt.baseImports, tt.fields, tt.index)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}
