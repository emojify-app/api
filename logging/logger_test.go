package logging

import (
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	l, _ := NewLogger("localhost:8323", "api", "create")

	defer l.CreateCalled("abc123", "http://something").Finished()
	time.Sleep(1 * time.Millisecond)
}
