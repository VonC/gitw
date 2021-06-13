package async

import (
	"gitw/internal/choices"
	"time"
)

type initAsync struct {
	elapsed  time.Duration
	timeout  time.Duration
	interval time.Duration
	choices.ChoicesManager
}

func NewInitAsync(interval, timeout time.Duration, cm choices.ChoicesManager) *initAsync {
	return &initAsync{
		elapsed:        0,
		timeout:        timeout,
		interval:       interval,
		ChoicesManager: cm,
	}
}

func (na *initAsync) IsActive() bool {
	return false
}

func (na *initAsync) PlaceHolder() string {
	return ""
}
func (na *initAsync) Increment() {}
func (na *initAsync) HasTimedOut() bool {
	return true
}
func (na *initAsync) Interval() time.Duration {
	return 500
}
