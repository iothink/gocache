package store

import (
	"fmt"
	"math"

	"testing"
)

// run go test -bench='BenchmarkPegasusStore*' -benchtime=1s -count=1 -run=none
func BenchmarkPegasusStore_Set(b *testing.B) {
	p, _ := NewPegasus(testPegasusOptions())
	defer p.Close()

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				p.Set(key, value, &Options{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkPegasusStore_Get(b *testing.B) {
	p, _ := NewPegasus(testPegasusOptions())
	defer p.Close()

	key := "test"
	value := []byte("value")

	p.Set(key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = p.Get(key)
			}
		})
	}
}
