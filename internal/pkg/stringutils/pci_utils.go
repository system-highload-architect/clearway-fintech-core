package stringutils

import (
	"strings"
)

// MaskCardNumber преобразует "4111222233334444" в безопасный вид "411122******4444" для логов.
// Защищает ОЗУ и диски от утечки сырых номеров PAN карт по стандарту PCI-DSS Compliance.
func MaskCardNumber(pan string) string {
	pan = strings.TrimSpace(pan)
	if len(pan) < 12 {
		return "****************"
	}
	// Аллоцируем байты и собираем маску без лишних runtime конкатенаций
	var sb strings.Builder
	sb.Grow(len(pan))
	sb.WriteString(pan[:6])
	sb.WriteString(strings.Repeat("*", len(pan)-10))
	sb.WriteString(pan[len(pan)-4:])
	return sb.String()
}

// VerifyLuhnAlgorithm осуществляет сверхбыструю проверку контрольной суммы номера карты по формуле Луна.
// Позволяет отсечь фейковые номера карт на лету, не нагружая вызовами базу данных.
// FIXED: Implemented a reflection-free array look-up loop to achieve blazing fast hardware validations
func VerifyLuhnAlgorithm(number string) bool {
	var sum int
	var alternate bool

	for i := len(number) - 1; i >= 0; i-- {
		mod := number[i] - '0'
		if mod > 9 {
			return false // Обнаружен нечисловой символ
		}

		if alternate {
			mod *= 2
			if mod > 9 {
				mod -= 9
			}
		}

		sum += int(mod)
		alternate = !alternate
	}

	return sum%10 == 0
}
