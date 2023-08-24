package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesHost(t *testing.T) {
	type args struct {
		host  string
		port  string
		host2 string
		port2 string
	}
	tests := []struct {
		name    string
		args    args
		matches bool
	}{
		{
			args: args{
				host:  "example.com",
				port:  "22",
				host2: "example.com",
				port2: "22",
			},
			matches: true,
		},
		{
			args: args{
				host:  "example.com",
				port:  "22",
				host2: "example.org",
				port2: "22",
			},
			matches: false,
		},
		{
			args: args{
				host:  "example.com",
				port:  "22",
				host2: "example.com",
				port2: "2222",
			},
			matches: false,
		},
		{
			args: args{
				host:  "login.example.com",
				port:  "22",
				host2: "*.example.com",
				port2: "22",
			},
			matches: true,
		},
		{
			args: args{
				host:  "example.com",
				port:  "22",
				host2: "*.example.com",
				port2: "22",
			},
			matches: false,
		},
	}

	for _, tt := range tests {
		assert.Equal(
			t,
			tt.matches,
			MatchesHost(tt.args.host, tt.args.port, tt.args.host2, tt.args.port2),
		)
	}
}

func TestGetenvs(t *testing.T) {
	keys := []string{"TEST_1", "TEST_2"}

	assert.Equal(t, Getenvs(keys...), "")

	os.Setenv(keys[1], keys[1])
	assert.Equal(t, Getenvs(keys...), keys[1])

	os.Setenv(keys[0], keys[0])
	assert.Equal(t, Getenvs(keys...), keys[0])
}
