package op

import (
	"reflect"
	"testing"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		input string
		uri   *URI
	}{
		{
			"op://app-prod/db/password",
			&URI{
				Vault: "app-prod",
				Item:  "db",
				Field: "password",
			},
		},
		{
			"op://app-prod/server/ssh/key.pem",
			&URI{
				Vault:   "app-prod",
				Item:    "server",
				Section: "ssh",
				Field:   "key.pem",
			},
		},
	}

	for _, tt := range tests {
		uri, err := ParseURI(tt.input)
		if err != nil {
			t.Errorf("ParseURI(%q) returned an error: %s", tt.input, err)
		}
		if !reflect.DeepEqual(uri, tt.uri) {
			t.Errorf("ParseURI(%q) returned %v, want %v", tt.input, uri, tt.uri)
		}
	}
}
