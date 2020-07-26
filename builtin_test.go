package bedis

import (
	"fmt"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	builtin, err := New(Option{
		Memory: "1mb",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer builtin.StopAndClose()

	client, err := builtin.DefaultClient()
	if err != nil {
		t.Fatal(err)
	}

	v := fmt.Sprint(time.Now().UnixNano())

	cmd := client.Set("key", v, -1)
	if err := cmd.Err(); err != nil {
		t.Fatal(err)
	}
	result, err := client.Get("key").Result()
	if err != nil {
		t.Fatalf("get err: %s", err)
	}
	if result != v {
		t.Fatalf("??? `%s` != `%s`", result, v)
	}

	// test LFU
	v += v + v + v
	for i := 0; i < 1e4; i++ {
		key := fmt.Sprintf("key-%d", i)
		cmd := client.Set(key, v, -1)
		if err := cmd.Err(); err != nil {
			t.Fatal(err)
		}
	}

	result, err = client.Get("key-10").Result()
	t.Logf("result: %s, err: %v", result, err)
}
