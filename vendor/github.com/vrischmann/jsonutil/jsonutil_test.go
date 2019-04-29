package jsonutil_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vrischmann/jsonutil"
)

func TestFromDuration(t *testing.T) {
	dur := 10 * time.Second
	require.Equal(t, jsonutil.FromDuration(dur).Duration, dur)
}

func TestMarshalJSON(t *testing.T) {
	d := jsonutil.Duration{time.Second * 10}
	data, err := d.MarshalJSON()

	require.Nil(t, err)
	require.Equal(t, `"10s"`, string(data))
}

func TestUnmarshalJSON(t *testing.T) {
	s := `"1m"`
	var d jsonutil.Duration
	err := d.UnmarshalJSON([]byte(s))

	require.Nil(t, err)
	require.Equal(t, time.Minute*1, d.Duration)
}
