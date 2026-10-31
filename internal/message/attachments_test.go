package message

import (
	"net/mail"
	"strings"
	"testing"
)

func parseTestMessage(t *testing.T, raw string) *mail.Message {
	t.Helper()

	msg, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("mail.ReadMessage() error = %v", err)
	}
	return msg
}

func TestExtractPDFAttachments(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []Attachment
		wantErr bool
	}{
		{
			name: "reads base64 PDF attachment",
			raw: "MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed; boundary=mail-boundary\r\n\r\n" +
				"--mail-boundary\r\n" +
				"Content-Type: application/pdf\r\n" +
				"Content-Disposition: attachment; filename=report.pdf\r\n" +
				"Content-Transfer-Encoding: base64\r\n\r\n" +
				"UERG\r\n" +
				"--mail-boundary--\r\n",
			want: []Attachment{{Name: "report.pdf", Content: []byte("PDF")}},
		},
		{
			name: "skips inline PDF and non-PDF attachment",
			raw: "MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed; boundary=mail-boundary\r\n\r\n" +
				"--mail-boundary\r\n" +
				"Content-Type: application/pdf\r\n" +
				"Content-Disposition: inline; filename=inline.pdf\r\n\r\n" +
				"inline bytes\r\n" +
				"--mail-boundary\r\n" +
				"Content-Type: application/msword\r\n" +
				"Content-Disposition: attachment; filename=report.doc\r\n\r\n" +
				"document bytes\r\n" +
				"--mail-boundary--\r\n",
		},
		{
			name: "uses Content-Type name fallback",
			raw: "MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/mixed; boundary=mail-boundary\r\n\r\n" +
				"--mail-boundary\r\n" +
				"Content-Type: application/pdf; name=report.pdf\r\n\r\n" +
				"PDF bytes\r\n" +
				"--mail-boundary--\r\n",
			want: []Attachment{{Name: "report.pdf", Content: []byte("PDF bytes")}},
		},
		{
			name: "no Content-Type returns no attachments",
			raw:  "From: sender@example.com\r\nSubject: no content type\r\n\r\nbody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := parseTestMessage(t, tt.raw)
			got, err := ExtractPDFAttachments(msg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractPDFAttachments() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ExtractPDFAttachments() returned %d attachments, want %d: %#v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i].Name != tt.want[i].Name || string(got[i].Content) != string(tt.want[i].Content) {
					t.Errorf("attachment[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// multipartTestBody собирает исходное MIME-тело независимо от кода извлечения,
// чтобы тесты проверяли разбор границ и кодировок на уровне формата письма.
func multipartTestBody(boundary string, parts ...string) string {
	var body strings.Builder
	for _, part := range parts {
		body.WriteString("--" + boundary + "\r\n")
		body.WriteString(part)
		body.WriteString("\r\n")
	}
	body.WriteString("--" + boundary + "--\r\n")
	return body.String()
}

func TestExtractPDFAttachmentsMIMEScenarios(t *testing.T) {
	message := func(parts ...string) string {
		return "Content-Type: multipart/mixed; boundary=outer\r\n\r\n" + multipartTestBody("outer", parts...)
	}
	pdf := func(name, content string) string {
		return "Content-Type: application/pdf\r\nContent-Disposition: attachment; filename=" + name + "\r\n\r\n" + content
	}
	nested := "Content-Type: multipart/mixed; boundary=middle\r\n\r\n" +
		multipartTestBody("middle",
			"Content-Type: text/plain\r\n\r\nmessage body",
			"Content-Type: multipart/related; boundary=inner\r\n\r\n"+
				multipartTestBody("inner", pdf("nested.pdf", "nested bytes")),
		)

	tests := []struct {
		name    string
		raw     string
		want    []Attachment
		wantErr bool
	}{
		{
			name: "preserves attachments before inside and after nested multipart",
			raw: message(pdf("before.pdf", "before"), nested,
				// Имя файла может быть задано и у части без Content-Type.
				"Content-Disposition: attachment; filename=after.pdf\r\n\r\nafter"),
			want: []Attachment{
				{Name: "before.pdf", Content: []byte("before")},
				{Name: "nested.pdf", Content: []byte("nested bytes")},
				{Name: "after.pdf", Content: []byte("after")},
			},
		},
		{
			name: "skips entire inline multipart and continues with next attachment",
			raw:  message("Content-Disposition: INLINE\r\n"+nested, pdf("kept.pdf", "kept")),
			want: []Attachment{{Name: "kept.pdf", Content: []byte("kept")}},
		},
		{
			name: "filename takes priority over Content-Type name",
			raw: message("Content-Type: application/octet-stream; name=wrong.doc\r\n" +
				"Content-Disposition: attachment; filename=REPORT.PDF\r\n\r\nbytes"),
			want: []Attachment{{Name: "REPORT.PDF", Content: []byte("bytes")}},
		},
		{
			name: "fallback when disposition exists without filename",
			raw: message("Content-Type: application/pdf; name=fallback.pdf\r\n" +
				"Content-Disposition: attachment\r\n\r\nbytes"),
			want: []Attachment{{Name: "fallback.pdf", Content: []byte("bytes")}},
		},
		{
			name: "quoted-printable body is decoded exactly once",
			raw: message("Content-Disposition: attachment; filename=quoted.pdf\r\n" +
				"Content-Transfer-Encoding: quoted-printable\r\n\r\n%PDF=3Ddata=00=FF"),
			want: []Attachment{{Name: "quoted.pdf", Content: []byte("%PDF=data\x00\xff")}},
		},
		{
			name: "RFC2231 filename is decoded and preserved",
			raw:  message("Content-Disposition: attachment; filename*=utf-8''%D0%93%D0%9F.pdf\r\n\r\nbytes"),
			want: []Attachment{{Name: "ГП.pdf", Content: []byte("bytes")}},
		},
		{
			name: "base64 preserves binary bytes",
			raw: message("Content-Disposition: attachment; filename=binary.pdf\r\n" +
				"Content-Transfer-Encoding: BASE64\r\n\r\nAP+A\r\n"),
			want: []Attachment{{Name: "binary.pdf", Content: []byte{0, 255, 128}}},
		},
		{
			name: "does not accept pdf in middle of filename or unnamed PDF part",
			raw:  message(pdf("report.pdf.exe", "not PDF"), "Content-Type: application/pdf\r\n\r\nunnamed"),
		},
		{
			name: "plain text message",
			raw:  "Content-Type: text/plain\r\n\r\nbody",
		},
		{
			name:    "missing outer boundary returns error",
			raw:     "Content-Type: multipart/mixed\r\n\r\nbody",
			wantErr: true,
		},
		{
			name:    "missing nested boundary returns error",
			raw:     message("Content-Type: multipart/mixed\r\n\r\nbody"),
			wantErr: true,
		},
		{
			name: "invalid base64 returns error",
			raw: message("Content-Disposition: attachment; filename=broken.pdf\r\n" +
				"Content-Transfer-Encoding: base64\r\n\r\n%%%"),
			wantErr: true,
		},
		{
			name:    "malformed disposition returns error",
			raw:     message("Content-Disposition: attachment; filename=\"unterminated\r\n\r\nbytes"),
			wantErr: true,
		},
		{
			name:    "truncated multipart returns error",
			raw:     "Content-Type: multipart/mixed; boundary=outer\r\n\r\n--outer\r\n" + pdf("broken.pdf", "bytes"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractPDFAttachments(parseTestMessage(t, tt.raw))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractPDFAttachments() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d attachments, want %d: %#v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i].Name != tt.want[i].Name || string(got[i].Content) != string(tt.want[i].Content) {
					t.Errorf("attachment[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
