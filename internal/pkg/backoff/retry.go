package backoff

import (
	"crypto/rand"
	"math/big"
	"time"
)

// RetryWithJitter выполняет b2b-операцию с экспоненциальным шагом задержки и случайным шумом.
// FIXED: Integrated a crypto-safe random noise offset to mitigate backend database spike bottlenecks
func RetryWithJitter(attempts int, baseDelay time.Duration, maxDelay time.Duration, task func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = task(); err == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		// Вычисляем экспоненциальную задержку: baseDelay * 2^attempt
		delay := baseDelay * (1 << uint(i))
		if delay > maxDelay {
			delay = maxDelay
		}

		// Добавляем случайный Jitter (шум) до 30% от текущей задержки
		jitterMax := int64(delay / 3)
		if jitterMax > 0 {
			nBig, randErr := rand.Int(rand.Reader, big.NewInt(jitterMax))
			if randErr == nil {
				delay += time.Duration(nBig.Int64())
			}
		}

		time.Sleep(delay)
	}
	return err
}
