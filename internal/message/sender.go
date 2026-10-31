package message

import (
	"net/mail"
)

// ExtractSender извлекает один адрес из значения заголовка From,
// например из строки «Отдел ДМС <gp@example.com>».
// Некорректный заголовок или список из нескольких адресов возвращает ошибку.
func ExtractSender(rawFrom string) (string, error) {
	addr, err := mail.ParseAddress(rawFrom)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}
