package retry

type Policy struct {
	maxRetries int
}

func NewPolicy(maxRetries int) Policy {
	return Policy{maxRetries: maxRetries}
}

func (p Policy) CanRetry(attempt int) bool {
	return attempt < p.maxRetries
}

func (p Policy) NextAttempt(attempt int) int {
	return attempt + 1
}
