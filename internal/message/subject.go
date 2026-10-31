package message

import "mime"

// ExtractSubject декодирует MIME encoded-word в значении заголовка Subject.
// Обычный текст остаётся без изменений; передавать всё письмо не нужно.
func ExtractSubject(rawSubject string) (string, error) {
	subject, err := (&mime.WordDecoder{}).DecodeHeader(rawSubject)
	if err != nil {
		return "", err
	}
	return subject, nil
}
