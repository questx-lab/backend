package buffer

import "sync"

var bufferPool = sync.Pool{
	New: func() any {
		return &buffer{buf: make([]byte, 1024)}
	},
}

type buffer struct {
	buf []byte
}

func New() *buffer {
	return bufferPool.Get().(*buffer)
}

// AppendBytes writes a single byte into buffer.
func (b *buffer) AppendByte(bz byte) {
	b.buf = append(b.buf, bz)
}

// AppendBytes writes a byte array into buffer.
func (b *buffer) AppendBytes(bz []byte) {
	b.buf = append(b.buf, bz...)
}

// Bytes returns the underlying byte slice of buffer.
func (b *buffer) Bytes() []byte {
	return b.buf
}

// Free puts the buffer into pool again. DO NOT USE the buffer after calling
// this method.
func (b *buffer) Free() {
	// Clear underlying array but remain the capacity of buffer.
	b.buf = b.buf[:0]
	bufferPool.Put(b)
}
