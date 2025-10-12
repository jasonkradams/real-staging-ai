package config

import (
	"testing"
)

func TestDatabaseURL(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "success: IPv4 host",
			config: Config{
				DB: DB{
					Host:     "192.168.1.1",
					Port:     5432,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
					SSLMode:  "disable",
				},
			},
			want: "postgres://testuser:testpass@192.168.1.1:5432/testdb?sslmode=disable",
		},
		{
			name: "success: IPv6 host (properly bracketed)",
			config: Config{
				DB: DB{
					Host:     "::1",
					Port:     5432,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
					SSLMode:  "disable",
				},
			},
			want: "postgres://testuser:testpass@[::1]:5432/testdb?sslmode=disable",
		},
		{
			name: "success: hostname",
			config: Config{
				DB: DB{
					Host:     "localhost",
					Port:     5432,
					User:     "testuser",
					Password: "testpass",
					Database: "testdb",
					SSLMode:  "require",
				},
			},
			want: "postgres://testuser:testpass@localhost:5432/testdb?sslmode=require",
		},
		{
			name: "success: URL set (overrides components)",
			config: Config{
				DB: DB{
					URL:      "postgres://user:pass@override:1234/db?sslmode=verify-full",
					Host:     "ignored",
					Port:     5432,
					User:     "ignored",
					Password: "ignored",
					Database: "ignored",
					SSLMode:  "ignored",
				},
			},
			want: "postgres://user:pass@override:1234/db?sslmode=verify-full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.DatabaseURL()
			if got != tt.want {
				t.Errorf("DatabaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
