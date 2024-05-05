package config

import (
	"net/url"
	"reflect"
	"testing"
)

func TestDatabaseConfig_ConnectionURL(t *testing.T) {
	type fields struct {
		User     string
		Password string
		Hostname string
		Port     string
		Name     string
	}
	tests := []struct {
		name   string
		fields fields
		want   url.URL
	}{
		{
			name: "no port",
			fields: fields{
				User:     "user",
				Password: "password",
				Hostname: "hostname",
				Name:     "name",
			},
			want: url.URL{
				Scheme: "postgres",
				User:   url.UserPassword("user", "password"),
				Host:   "hostname",
				Path:   "name",
			},
		},
		{
			name: "with port",
			fields: fields{
				User:     "user",
				Password: "password",
				Hostname: "hostname",
				Port:     "5432",
				Name:     "name",
			},
			want: url.URL{
				Scheme: "postgres",
				User:   url.UserPassword("user", "password"),
				Host:   "hostname:5432",
				Path:   "name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &DatabaseConfig{
				User:     tt.fields.User,
				Password: tt.fields.Password,
				Hostname: tt.fields.Hostname,
				Port:     tt.fields.Port,
				Name:     tt.fields.Name,
			}
			if got := c.ConnectionURL(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseConfig.ConnectionURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
