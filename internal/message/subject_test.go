package message

import "testing"

func TestExtractSubject(t *testing.T) {
	tests := []struct {
		name       string
		rawSubject string
		want       string
	}{
		{name: "plain subject", rawSubject: "ГП — направление", want: "ГП — направление"},
		{name: "encoded UTF-8 subject", rawSubject: "=?UTF-8?B?0JPQnw==?=", want: "ГП"},
		{name: "encoded and plain text", rawSubject: "=?UTF-8?B?0JPQnw==?= — направление", want: "ГП — направление"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSubject(tt.rawSubject)
			if err != nil {
				t.Fatalf("ExtractSubject(%q) error = %v", tt.rawSubject, err)
			}
			if got != tt.want {
				t.Errorf("ExtractSubject(%q) = %q, want %q", tt.rawSubject, got, tt.want)
			}
		})
	}
}
