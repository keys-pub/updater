package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeToMillis(t *testing.T) {
	t1 := time.Now()
	ts := TimeToMillis(t1)
	t2 := TimeFromMillis(ts)
	require.Equal(t, t1.Format(time.StampMilli), t2.Format(time.StampMilli))

	require.Equal(t, TimeMs(0), TimeToMillis(time.Time{}))
	require.Equal(t, time.Time{}, TimeFromMillis(0))
}
