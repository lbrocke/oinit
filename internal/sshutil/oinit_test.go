package sshutil

import (
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
			matchesHost(tt.args.host, tt.args.port, tt.args.host2, tt.args.port2),
		)
	}
}
