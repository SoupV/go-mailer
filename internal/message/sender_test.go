package message

import "testing"

func TestExtractSender(t *testing.T) {
	tests := []struct {
		name    string
		rawFrom string
		want    string
		wantErr bool
	}{
		{name: "bare address", rawFrom: "gp@renins.com", want: "gp@renins.com"},
		{name: "display name", rawFrom: "Отдел ДМС <gp@renins.com>", want: "gp@renins.com"},
		{name: "invalid address", rawFrom: "not-an-email", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSender(tt.rawFrom)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractSender(%q) error = %v, wantErr %t", tt.rawFrom, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("ExtractSender(%q) = %q, want %q", tt.rawFrom, got, tt.want)
			}
		})
	}
}
