package cachify

import (
	"time"
)

func NewEntries() *entries {
	return &entries{}
}

func NewState() *state {
	return &state{
		accessTime: time.Now(),
	}
}

func (c *entries) WithKey(value string) *entries {
	c.key = value
	return c
}

func (c *entries) WithValue(value interface{}) *entries {
	c.value = value
	return c
}

func (c *entries) WithExpiration(value time.Time) *entries {
	c.expiration = value
	return c
}

func (c *entries) Key() string {
	return c.key
}

func (c *entries) Value() interface{} {
	return c.value
}

func (c *entries) Expiration() time.Time {
	return c.expiration
}

func (l *state) WithKey(value string) *state {
	l.key = value
	return l
}

func (l *state) WithValue(value interface{}) *state {
	l.value = value
	return l
}

func (l *state) WithAccessTime(value time.Time) *state {
	l.accessTime = value
	return l
}

func (l *state) WithExpiration(value time.Time) *state {
	l.expiration = value
	return l
}

func (l *state) Key() string {
	return l.key
}

func (l *state) Value() interface{} {
	return l.value
}

func (l *state) Expiration() time.Time {
	return l.expiration
}

func (l *state) AccessTime() time.Time {
	return l.accessTime
}
