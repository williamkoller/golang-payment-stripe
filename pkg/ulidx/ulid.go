package ulidx

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func New() string {
	t := ulid.Timestamp(time.Now())
	return ulid.MustNew(t, ulid.Monotonic(rand.Reader, 0)).String()
}
