package async

import "time"

type NoAsync struct{}

func (na *NoAsync) IsActive() bool {
	return false
}

func (na *NoAsync) CachedChoices() []string {
	return nil
}
func (na *NoAsync) RetrievedChoices() []string {
	return nil
}
func (na *NoAsync) PlaceHolder() string {
	return ""
}
func (na *NoAsync) Increment() {}
func (na *NoAsync) HasTimedOut() bool {
	return true
}
func (na *NoAsync) Interval() time.Duration {
	return 0
}
