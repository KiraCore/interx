package log

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// Recovery function to handle panics
func RecoverFromPanic() {
	if r := recover(); r != nil {
		logger.Errorf("Application crashed: %v", r)
	}
}
