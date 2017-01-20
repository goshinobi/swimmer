package queue

import (
	"net/url"
	"time"
)

type Queue struct {
	queue       map[string]chan url.URL
	lastPopTime time.Time
	delayTime   int64
	size        int
}

func NewQueue(size int, delayTime int64) *Queue {
	return &Queue{
		size:      size,
		delayTime: delayTime,
	}
}
