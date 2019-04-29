package jsonutil_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/vrischmann/jsonutil"
)

func ExampleMarshalJSON() {
	d := jsonutil.Duration{time.Minute * 10}

	data, _ := json.Marshal(d)
	fmt.Println(string(data))
	// Output:
	// "10m0s"
}

func ExampleUnmarshalJSON() {
	var d jsonutil.Duration
	s := `"10m"`

	_ = json.Unmarshal([]byte(s), &d)
	fmt.Println(d.Duration.Seconds())
	// Output:
	// 600
}
