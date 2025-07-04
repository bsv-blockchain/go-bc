package bc

import (
	"encoding/binary"
	"testing"
)

// FuzzUInt32ToBytes ensures UInt32ToBytes correctly round-trips values.
// The function converts a uint32 to little-endian bytes.  Fuzzing verifies
// that converting back using the standard library yields the original value.
func FuzzUInt32ToBytes(f *testing.F) {
	f.Add(uint32(0))
	f.Add(uint32(1))
	f.Add(uint32(123456))
	f.Add(^uint32(0))

	f.Fuzz(func(t *testing.T, v uint32) {
		b := UInt32ToBytes(v)
		if len(b) != 4 {
			t.Fatalf("unexpected length %d", len(b))
		}
		round := binary.LittleEndian.Uint32(b)
		if round != v {
			t.Fatalf("round trip failed: %d != %d", v, round)
		}
	})
}
