package sipproxy

import "testing"

func TestSIPWildcardListenAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		port int
		want string
	}{
		{name: "sip default port", port: 5060, want: ":5060"},
		{name: "ephemeral port", port: 0, want: ":0"},
		{name: "invalid port uses ephemeral", port: -1, want: ":0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sipWildcardListenAddr(tt.port)
			if got != tt.want {
				t.Fatalf("sipWildcardListenAddr(%d) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}
