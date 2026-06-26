package database

import (
	"strings"
	"testing"
)

func TestDSN(t *testing.T) {
	tests := []struct {
		name     string
		conf     DBConf
		contains []string
	}{
		{
			"default local", DBConf{Host: "127.0.0.1", Port: 3306, User: "root", Password: "s", Database: "db", MaxConns: 20, MaxIdle: 10},
			[]string{"root:s@tcp(127.0.0.1:3306)/db", "charset=utf8mb4", "parseTime=True", "loc=Local"},
		},
		{
			"remote host", DBConf{Host: "192.168.211.58", Port: 3306, User: "admin", Password: "Abc123654", Database: "business"},
			[]string{"admin:Abc123654@tcp(192.168.211.58:3306)/business"},
		},
		{
			"non-standard port", DBConf{Host: "db.internal", Port: 3307, User: "app", Password: "pwd", Database: "prod"},
			[]string{"app:pwd@tcp(db.internal:3307)/prod"},
		},
		{
			"empty password", DBConf{Host: "localhost", Port: 3306, User: "root", Password: "", Database: "test"},
			[]string{"root:@tcp(localhost:3306)/test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.conf.DSN()
			for _, exp := range tt.contains {
				if !strings.Contains(dsn, exp) {
					t.Errorf("DSN = %s, want containing %s", dsn, exp)
				}
			}
		})
	}
}
