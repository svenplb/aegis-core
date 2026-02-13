package redactor

import "fmt"

// Counter assigns incrementing placeholder tokens per entity type.
// If the same original text is seen again, the previously assigned token is reused.
type Counter struct {
	counts map[string]int
	seen   map[string]string // original text → token
}

// NewCounter returns a ready-to-use Counter.
func NewCounter() *Counter {
	return &Counter{
		counts: make(map[string]int),
		seen:   make(map[string]string),
	}
}

// Next returns a placeholder token for the given entity type and original text.
// Repeated calls with the same originalText return the same token.
func (c *Counter) Next(entityType, originalText string) string {
	if tok, ok := c.seen[originalText]; ok {
		return tok
	}
	c.counts[entityType]++
	tok := fmt.Sprintf("[%s_%d]", entityType, c.counts[entityType])
	c.seen[originalText] = tok
	return tok
}
