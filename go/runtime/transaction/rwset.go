package transaction

import (
	"bytes"

	"github.com/oasislabs/ekiden/go/common/cbor"
)

var (
	_ cbor.Marshaler   = (*ReadWriteSet)(nil)
	_ cbor.Unmarshaler = (*ReadWriteSet)(nil)
)

// CoarsenedKey is a coarsened key prefix that represents any key that
// starts with this prefix.
type CoarsenedKey []byte

// CoarsenedSet is a set of coarsened keys.
type CoarsenedSet []CoarsenedKey

// Equal checks whether the coarsened set is equal to another.
func (s CoarsenedSet) Equal(other CoarsenedSet) bool {
	if len(s) != len(other) {
		return false
	}
	for idx, value := range s {
		if !bytes.Equal(value, other[idx]) {
			return false
		}
	}
	return true
}

// ReadWriteSet is a read/write set.
type ReadWriteSet struct {
	// Granularity is the size of the key prefixes (in bytes) used for
	// coarsening the keys.
	Granularity uint16 `codec:"granularity"`
	// ReadSet is the read set.
	ReadSet CoarsenedSet `codec:"read_set"`
	// WriteSet is the write set.
	WriteSet CoarsenedSet `codec:"write_set"`
}

// MarshalCBOR serializes the type into a CBOR byte vector.
func (rw *ReadWriteSet) MarshalCBOR() []byte {
	return cbor.Marshal(rw)
}

// UnmarshalCBOR deserializes a CBOR byte vector into given type.
func (rw *ReadWriteSet) UnmarshalCBOR(data []byte) error {
	return cbor.Unmarshal(data, rw)
}

// Equal checks whether the read/write set is equal to another.
func (rw ReadWriteSet) Equal(other *ReadWriteSet) bool {
	return rw.Granularity == other.Granularity && rw.ReadSet.Equal(other.ReadSet) && rw.WriteSet.Equal(other.WriteSet)
}

// Size returns the size (in bytes) of the read/write set.
func (rw ReadWriteSet) Size() uint64 {
	return uint64(len(rw.MarshalCBOR()))
}