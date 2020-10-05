package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	ch := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	ch.Add("2", "4", "6")

	testCases := map[string]string{
		"2": "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, exp := range testCases {
		if v := ch.Get(k); v != exp {
			t.Fatalf("get error node of key : %s, node: %s, expect node: %s", k, v, exp)
		}
	}

	// 2, 4, 6, 8, 12, 14, 16, 18, 22, 24, 26, 28
	ch.Add("8")

	testCases["27"] = "8"

	for k, exp := range testCases {
		if v := ch.Get(k); v != exp {
			t.Fatalf("get error node of key : %s, node: %s, expect node: %s", k, v, exp)
		}
	}
}