package rfc8888

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetArrivalTimeOffset(t *testing.T) {
	for _, test := range []struct {
		base    time.Time
		arrival time.Time
		want    uint16
	}{
		{
			base:    time.Time{}.Add(time.Second),
			arrival: time.Time{},
			want:    1024,
		},
		{
			base:    time.Time{}.Add(500 * time.Millisecond),
			arrival: time.Time{},
			want:    512,
		},
		{
			base:    time.Time{}.Add(8 * time.Second),
			arrival: time.Time{},
			want:    0x1FFE,
		},
		{
			base:    time.Time{},
			arrival: time.Time{}.Add(time.Second),
			want:    0x1FFF,
		},
	} {
		assert.Equal(t, test.want, getArrivalTimeOffset(test.base, test.arrival))
	}
}

func TestAddPacket(t *testing.T) {

}

func TestBuildReport(t *testing.T) {
}
