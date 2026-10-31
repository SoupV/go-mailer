// Package message разбирает заголовки писем и извлекает PDF из MIME-частей.
package message

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"path/filepath"
	"strings"
)

// Attachment хранит имя вложения и его байты после декодирования передачи.
type Attachment struct {
	Name    string
	Content []byte
}

// ExtractPDFAttachments извлекает именованные PDF из multipart-письма,
// включая вложенные multipart-контейнеры, и пропускает inline-части.
// Отбор выполняется по расширению .pdf без учёта регистра.
// Функция читает msg.Body и сохраняет содержимое найденных файлов в памяти.
func ExtractPDFAttachments(msg *mail.Message) ([]Attachment, error) {
	contentHeader := msg.Header.Get("Content-Type")
	if contentHeader == "" {
		// При отсутствии MIME-типа считаем письмо обычным текстовым.
		return nil, nil
	}

	mediaType, params, err := mime.ParseMediaType(contentHeader)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		// Извлечение из одиночного MIME-тела пока не поддерживается.
		return nil, nil
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, fmt.Errorf("у multipart письма отсутствует boundary")
	}

	reader := multipart.NewReader(msg.Body, boundary)
	var attachments []Attachment
	attachments, err = extractPDFParts(reader)
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

// extractPDFParts обходит один multipart-уровень и рекурсивно читает вложенные.
func extractPDFParts(reader *multipart.Reader) ([]Attachment, error) {
	var attachments []Attachment

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return attachments, nil
		}
		if err != nil {
			return nil, err
		}

		filename := ""

		if header := part.Header.Get("Content-Disposition"); header != "" {
			disp, params, err := mime.ParseMediaType(header)
			if err != nil {
				return nil, err
			}

			if strings.EqualFold(disp, "inline") {
				// Пропускаем и отдельные inline-файлы, и целые inline-контейнеры.
				continue
			}
			filename = params["filename"]
		}

		// NextPart обходит только текущий уровень. У вложенного контейнера
		// собственная boundary, а сама part служит reader для его тела.
		contentHeader := part.Header.Get("Content-Type")
		if contentHeader != "" {
			mediaType, params, err := mime.ParseMediaType(contentHeader)
			if err != nil {
				return nil, err
			}

			if strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
				nestedBoundary := params["boundary"]
				if nestedBoundary == "" {
					return nil, fmt.Errorf("у multipart письма отсутствует boundary")
				}
				nestedAttachments, err := extractPDFParts(
					multipart.NewReader(part, nestedBoundary),
				)
				if err != nil {
					return nil, err
				}
				attachments = append(attachments, nestedAttachments...)
				continue
			}
		}

		if filename == "" {
			// Некоторые отправители указывают имя только в Content-Type: name=...
			if header := part.Header.Get("Content-Type"); header != "" {
				_, params, err := mime.ParseMediaType(header)
				if err != nil {
					return nil, err
				}
				filename = params["name"]
			}
		}

		if filename == "" || !strings.EqualFold(filepath.Ext(filename), ".pdf") {
			continue
		}

		encoding := strings.TrimSpace(part.Header.Get("Content-Transfer-Encoding"))

		var source io.Reader = part

		// NextPart уже декодирует quoted-printable, но base64 нужно декодировать явно.
		if strings.EqualFold(encoding, "base64") {
			source = base64.NewDecoder(base64.StdEncoding, part)
		}

		content, err := io.ReadAll(source)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, Attachment{
			Name:    filename,
			Content: content,
		})
	}
}
