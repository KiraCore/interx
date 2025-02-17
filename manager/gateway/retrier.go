package gateway

import (
	"errors"
	"github.com/saiset-co/sai-interx-manager/logger"
	"go.uber.org/zap"
	"time"
)

type RetryFunc func() (interface{}, error)

type Retrier struct {
	attempts int
	delay    time.Duration
}

func NewRetrier(attempts int, delay time.Duration) *Retrier {
	return &Retrier{
		attempts: attempts,
		delay:    delay,
	}
}

func (r *Retrier) Do(fn RetryFunc) (interface{}, error) {
	for i := 0; i < r.attempts; i++ {
		result, err := fn()
		if err == nil {
			logger.Logger.Error("Retrier - Do", zap.Error(err))
			return result, nil
		}

		if i < r.attempts-1 { // Don't sleep after the last attempt
			time.Sleep(r.delay)
		}
	}

	err := errors.New("all retry attempts failed")
	logger.Logger.Error("Retrier - Do", zap.Error(err))

	return nil, err
}
