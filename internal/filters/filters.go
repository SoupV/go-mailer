// Package filters содержит правила отбора писем по теме, отправителю и именам вложений.
package filters

import (
	"path/filepath"
	"strings"
)

// SubjectMatches ищет хотя бы одну непустую фразу в уже декодированной теме.
// Поиск регистрозависимый, без проверки границ слов.
func SubjectMatches(subject string, keywords []string) bool {
	for _, keyword := range keywords {
		// Пустая подстрока совпадает с любой темой, поэтому её игнорируем.
		if keyword == "" {
			continue
		}
		if strings.Contains(subject, keyword) {
			return true
		}
	}
	return false
}

// HasPDFAttachment проверяет расширения имён без учёта регистра.
// Содержимое файлов и их MIME-типы здесь не проверяются.
func HasPDFAttachment(files []string) bool {
	for _, file := range files {
		if strings.EqualFold(filepath.Ext(file), ".pdf") {
			return true
		}
	}
	return false
}

// SenderAllowed ищет точное совпадение адреса без учёта регистра и пробелов по краям.
// sender должен содержать сам адрес, уже извлечённый из заголовка From.
func SenderAllowed(sender string, allowedSenders []string) bool {
	sender = strings.TrimSpace(sender)
	if sender == "" {
		return false
	}
	for _, allowedSender := range allowedSenders {
		allowedSender = strings.TrimSpace(allowedSender)
		if allowedSender == "" {
			continue
		}
		if strings.EqualFold(sender, allowedSender) {
			return true
		}
	}
	return false
}
