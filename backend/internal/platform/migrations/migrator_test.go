package migrations

import "testing"

func TestMigrationVersion(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     int64
		wantErr  bool
	}{
		{name: "valid", filename: "00001_create_users.sql", want: 1},
		{name: "large version", filename: "42_add_index.sql", want: 42},
		{name: "missing separator", filename: "users.sql", wantErr: true},
		{name: "non numeric version", filename: "first_users.sql", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := migrationVersion(test.filename)
			if test.wantErr {
				if err == nil {
					t.Fatalf("migrationVersion(%q) returned no error", test.filename)
				}
				return
			}
			if err != nil {
				t.Fatalf("migrationVersion(%q) returned error: %v", test.filename, err)
			}
			if got != test.want {
				t.Fatalf("migrationVersion(%q) = %d, want %d", test.filename, got, test.want)
			}
		})
	}
}
