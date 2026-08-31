package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeEntityFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "entity.go")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func Test_entityConstructorInfo(t *testing.T) {
	tests := []struct {
		name       string
		entityName string
		content    string
		wantErr    bool
		wantArgs   string
		wantFields []Field
		wantUUID   bool
		wantTime   bool
	}{
		{
			name:       "string id plus extra field",
			entityName: "Category",
			content:    "package categories\nfunc NewCategory(id string, description string) (*Category, error) { return nil, nil }\n",
			wantArgs:   `uuid.NewString(), input.Description`,
			wantFields: []Field{{Name: "Description", Type: "string"}},
			wantUUID:   true,
		},
		{
			name:       "uuid id and timestamp",
			entityName: "User",
			content:    "package users\nfunc NewUser(id uuid.UUID, createdAt time.Time) (*User, error) { return nil, nil }\n",
			wantArgs:   "uuid.New(), time.Now()",
			wantUUID:   true,
			wantTime:   true,
		},
		{
			name:       "id only",
			entityName: "Note",
			content:    "package notes\nfunc NewNote(id string) (*Note, error) { return nil, nil }\n",
			wantArgs:   "uuid.NewString()",
			wantUUID:   true,
		},
		{
			name:       "no id field",
			entityName: "Greeting",
			content:    "package greetings\nfunc NewGreeting(name string, email string) (*Greeting, error) { return nil, nil }\n",
			wantArgs:   "input.Name, input.Email",
			wantFields: []Field{{Name: "Name", Type: "string"}, {Name: "Email", Type: "string"}},
		},
		{
			name:       "grouped param names",
			entityName: "Pair",
			content:    "package pairs\nfunc NewPair(a, b string) (*Pair, error) { return nil, nil }\n",
			wantArgs:   "input.A, input.B",
			wantFields: []Field{{Name: "A", Type: "string"}, {Name: "B", Type: "string"}},
		},
		{
			name:       "pointer and slice fields become input fields",
			entityName: "Order",
			content:    "package orders\nfunc NewOrder(id string, total float64, items []string, at *time.Time) (*Order, error) { return nil, nil }\n",
			wantArgs:   "uuid.NewString(), input.Total, input.Items, input.At",
			wantFields: []Field{
				{Name: "Total", Type: "float64"},
				{Name: "Items", Type: "[]string"},
				{Name: "At", Type: "*time.Time"},
			},
			wantUUID: true,
			wantTime: true,
		},
		{
			name:       "constructor missing",
			entityName: "Category",
			content:    "package categories\n",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			path := writeEntityFile(t, tt.content)
			info, err := entityConstructorInfo(path, tt.entityName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantArgs, info.newArgs)
			assert.Equal(t, tt.wantFields, info.createFields)
			assert.Equal(t, tt.wantUUID, info.needsUUID)
			assert.Equal(t, tt.wantTime, info.needsTime)
		})
	}
}

func Test_entityConstructorInfo_missingFile(t *testing.T) {
	_, err := entityConstructorInfo(filepath.Join(t.TempDir(), "nope.go"), "Category")
	assert.Error(t, err)
}
