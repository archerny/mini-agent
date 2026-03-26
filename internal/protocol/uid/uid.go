// Package uid provides time-ordered unique ID generation.
//
// We use UUID v7 (RFC 9562) which embeds a Unix timestamp in the high bits,
// giving us both uniqueness and natural time ordering — ideal for messages
// and events that need to be sorted chronologically.
package uid

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	lastMS  int64
	counter uint16
)

// New generates a UUID v7 string.
//
// Format: xxxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx
//
//   - Bits 0-47:  Unix timestamp in milliseconds
//   - Bits 48-51: Version (0b0111 = 7)
//   - Bits 52-63: Sub-millisecond counter / random
//   - Bits 64-65: Variant (0b10)
//   - Bits 66-127: Random
func New() string {
	mu.Lock()

	ms := time.Now().UnixMilli()
	if ms == lastMS {
		counter++
	} else {
		lastMS = ms
		counter = 0
	}
	seq := counter

	mu.Unlock()

	var uuid [16]byte

	// Bytes 0-5: timestamp (48 bits, big-endian)
	binary.BigEndian.PutUint16(uuid[0:2], uint16(ms>>32))
	binary.BigEndian.PutUint32(uuid[2:6], uint32(ms))

	// Bytes 6-7: version (4 bits) + seq_hi (12 bits)
	binary.BigEndian.PutUint16(uuid[6:8], seq)
	uuid[6] = (uuid[6] & 0x0F) | 0x70 // version 7

	// Bytes 8-15: variant (2 bits) + random (62 bits)
	_, _ = rand.Read(uuid[8:16])
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // variant 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(uuid[0:4]),
		binary.BigEndian.Uint16(uuid[4:6]),
		binary.BigEndian.Uint16(uuid[6:8]),
		binary.BigEndian.Uint16(uuid[8:10]),
		uuid[10:16],
	)
}
