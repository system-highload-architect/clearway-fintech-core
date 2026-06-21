package fixedpoint

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrAmountNegative = errors.New("💸 [FINTECH MATH]: Сумма платежа не может быть отрицательной")
	ErrOverflow       = errors.New("💸 [FINTECH MATH]: Критическое переполнение разрядной сетки ОЗУ")
	ErrInvalidFormat  = errors.New("💸 [FINTECH MATH]: Невалидный формат строки денежной суммы")
)

// Money представляет точную финтех-сумму в минимальных неделимых единицах (копейках).
// Хранится в плоском int64, гарантируя 0 аллокаций памяти и вычисления на регистрах CPU.
type Money struct {
	Units int64 // Например: 125.50 рублей хранится как 12550 копеек
}

// NewMoneyFromInt64 создаёт объект напрямую из копеек (канонический b2b-путь)
func NewMoneyFromInt64(units int64) Money {
	return Money{Units: units}
}

// NewMoneyFromString осуществляет ручной побайтовый парсинг строки (без fmt.Sscanf и без float64).
// Работает со скоростью света, не нагружает Garbage Collector.
func NewMoneyFromString(val string) (Money, error) {
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return Money{}, ErrInvalidFormat
	}

	if val[0] == '-' {
		return Money{}, ErrAmountNegative
	}

	// Ищем разделитель дробной части (точку)
	dotIdx := strings.IndexByte(val, '.')

	var rublesStr, kopecksStr string
	if dotIdx == -1 {
		rublesStr = val
		kopecksStr = "00"
	} else {
		rublesStr = val[:dotIdx]
		kopecksStr = val[dotIdx+1:]

		// Нормализуем копейки до строго двух знаков (например, ".5" -> "50", ".532" -> "53")
		if len(kopecksStr) == 1 {
			kopecksStr += "0"
		} else if len(kopecksStr) > 2 {
			kopecksStr = kopecksStr[:2]
		}
	}

	// Превращаем рубли в int64 без аллокаций
	rubles, err := strconv.ParseInt(rublesStr, 10, 64)
	if err != nil {
		return Money{}, ErrInvalidFormat
	}

	// Превращаем копейки в int64
	kopecks, err := strconv.ParseInt(kopecksStr, 10, 64)
	if err != nil {
		return Money{}, ErrInvalidFormat
	}

	// Защита от переполнения при умножении рублей на 100
	if rubles > (1<<63-1)/100 {
		return Money{}, ErrOverflow
	}

	return Money{Units: rubles*100 + kopecks}, nil
}

// Add производит наносекундное сложение с защитой от переполнения разрядной сетки CPU
func (m Money) Add(other Money) (Money, error) {
	if other.Units > 0 && m.Units > (1<<63-1)-other.Units {
		return Money{}, ErrOverflow
	}
	return Money{Units: m.Units + other.Units}, nil
}

// Sub производит вычитание (проверка на овердрафт счета мерчанта)
func (m Money) Sub(other Money) (Money, error) {
	if m.Units < other.Units {
		return Money{}, errors.New("🔒 [LEDGER GUARD]: Недостаточно средств на балансе кошелька")
	}
	return Money{Units: m.Units - other.Units}, nil
}

// String собирает строку назад ("150.25") через strings.Builder с аллокацией буфера в 1 проход
// FIXED: Optimized via strings.Builder Grow API to bypass intermediate runtime allocations
func (m Money) String() string {
	rubles := m.Units / 100
	kopecks := m.Units % 100
	if kopecks < 0 {
		kopecks = -kopecks
	}

	rublesStr := strconv.FormatInt(rubles, 10)

	// Аллоцируем память строго под размер итоговой строки: длина рублей + точка + 2 знака копеек
	var sb strings.Builder
	sb.Grow(len(rublesStr) + 1 + 2)

	sb.WriteString(rublesStr)
	sb.WriteByte('.')
	if kopecks < 10 {
		sb.WriteByte('0')
	}
	sb.WriteString(strconv.FormatInt(kopecks, 10))

	return sb.String()
}
