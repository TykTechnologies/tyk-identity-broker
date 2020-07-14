package rendezvous

import (
	"hash/maphash"
	"testing"
)

var hseed = maphash.MakeSeed()

func hashString(s string) uint64 {
	var h maphash.Hash
	h.SetSeed(hseed)
	h.WriteString(s)
	return h.Sum64()
}

func TestEmpty(t *testing.T) {
	r := New([]string{}, hashString)
	r.Lookup("hello")

}
