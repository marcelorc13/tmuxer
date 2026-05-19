package tmux

import (
	"testing"
)

func TestParseSessions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Session
	}{
		{
			name: "empty output",
			in:   "",
			want: nil,
		},
		{
			name: "single detached session",
			in:   "main:0",
			want: []Session{{Name: "main", Attached: false}},
		},
		{
			name: "single attached session",
			in:   "work:1",
			want: []Session{{Name: "work", Attached: true}},
		},
		{
			name: "multiple sessions",
			in:   "main:0\nwork:1\nplay:0",
			want: []Session{
				{Name: "main", Attached: false},
				{Name: "work", Attached: true},
				{Name: "play", Attached: false},
			},
		},
		{
			name: "session name with colons",
			in:   "my:session:0",
			want: []Session{{Name: "my", Attached: false}},
		},
		{
			name: "trailing newline",
			in:   "main:0\n",
			want: []Session{{Name: "main", Attached: false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSessions(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d sessions, want %d", len(got), len(tt.want))
			}
			for i, s := range got {
				if s != tt.want[i] {
					t.Errorf("session[%d]: got %+v, want %+v", i, s, tt.want[i])
				}
			}
		})
	}
}
