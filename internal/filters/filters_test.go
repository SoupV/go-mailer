package filters

import "testing"

func TestSubjectMatches(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		keywords []string
		want     bool
	}{
		{name: "contains keyword", subject: "ГП — направление", keywords: []string{"ГП"}, want: true},
		{name: "case sensitive", subject: "гп — направление", keywords: []string{"ГП"}, want: false},
		{name: "any keyword can match", subject: "Гарантийное письмо", keywords: []string{"ГП", "Гарантийное письмо"}, want: true},
		{name: "empty keyword is ignored", subject: "Любая тема", keywords: []string{""}, want: false},
		{name: "no keywords", subject: "ГП — направление", keywords: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubjectMatches(tt.subject, tt.keywords); got != tt.want {
				t.Errorf("SubjectMatches(%q, %v) = %t, want %t", tt.subject, tt.keywords, got, tt.want)
			}
		})
	}
}

func TestHasPDFAttachment(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{name: "pdf extension", files: []string{"scan.pdf", "photo.jpg"}, want: true},
		{name: "uppercase extension", files: []string{"scan.PDF"}, want: true},
		{name: "pdf is not final extension", files: []string{"report.pdf.exe"}, want: false},
		{name: "other extensions", files: []string{"scan.docx", "photo.jpg"}, want: false},
		{name: "no attachments", files: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPDFAttachment(tt.files); got != tt.want {
				t.Errorf("HasPDFAttachment(%v) = %t, want %t", tt.files, got, tt.want)
			}
		})
	}
}

func TestSenderAllowed(t *testing.T) {
	tests := []struct {
		name           string
		sender         string
		allowedSenders []string
		want           bool
	}{
		{name: "exact address", sender: "gp@renins.com", allowedSenders: []string{"gp@renins.com"}, want: true},
		{name: "case insensitive", sender: "GP@renins.com", allowedSenders: []string{"gp@renins.com"}, want: true},
		{name: "not a substring match", sender: "other-gp@renins.com", allowedSenders: []string{"gp@renins.com"}, want: false},
		{name: "trims surrounding spaces", sender: " gp@renins.com ", allowedSenders: []string{" gp@renins.com "}, want: true},
		{name: "empty sender", sender: "", allowedSenders: []string{""}, want: false},
		{name: "whitespace sender", sender: " ", allowedSenders: []string{" "}, want: false},
		{name: "empty allowlist", sender: "gp@renins.com", allowedSenders: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SenderAllowed(tt.sender, tt.allowedSenders); got != tt.want {
				t.Errorf("SenderAllowed(%q, %v) = %t, want %t", tt.sender, tt.allowedSenders, got, tt.want)
			}
		})
	}
}
