package runebuffer

import (
	"bufio"
	"io"
	"sync"
)

const (
	DefaultBufferSize = 1024
)

// RuneBuffer adds an extra layer of buffering on top of bufio.Reader that works entirely with runes.
// This enables multiple UnreadRune calls without an intermediate read operation.
// This is a ring buffer, so the max number of UnreadRune calls will always be <= DefaultBufferSize (or the size passed to NewRuneBufferWithSize).
type RuneBuffer struct {
	br           *bufio.Reader
	rptr, wptr   int
	buf          []rune
	size, unread int
}

func NewRuneBuffer(r io.Reader) *RuneBuffer {
	return NewRuneBufferWithSize(r, DefaultBufferSize)
}

func NewRuneBufferWithSize(r io.Reader, size int) *RuneBuffer {
	if r == nil {
		return nil
	}
	return &RuneBuffer{
		br:  bufio.NewReader(r),
		buf: make([]rune, size),
	}
}

// ReadRune will read the next rune in the buffer, pulling from the io.Reader if necessary.
func (b *RuneBuffer) ReadRune() (rune, error) {
	if b.unread == 0 {
		if b.wptr == -1 {
			return 0, nil
		}
		if err := b.readStream(); err != nil {
			return 0, err
		}
	}
	val := b.buf[b.rptr]
	b.incrementRptr()
	b.decrementUnread()
	return val, nil
}

// UnreadRune will unread the previously read rune, if it exists.
// If no runes have been read, or the read pointer has reached the beginning of the buffer, this is a no-op.
func (b *RuneBuffer) UnreadRune() {
	b.unreadRunes(1)
}

// UnreadNumRunes will unread the specified number of runes.
func (b *RuneBuffer) UnreadNumRunes(num int) {
	b.unreadRunes(num)
}

func (b *RuneBuffer) unreadRunes(num int) {
	for i := 0; i < num; i++ {
		if b.size == 0 {
			return
		}
		if b.unread == b.size {
			return
		}
		b.incrementUnread()
		b.decrementRptr()
	}
}

func (b *RuneBuffer) readStream() error {
	r, _, err := b.br.ReadRune()
	if err != nil && err != io.EOF {
		return err
	}
	b.buf[b.wptr] = r
	if b.shouldPushRptr() {
		b.incrementRptr()
	} else {
		// Pushing the read pointer leaves the same amount in unread as before.
		b.incrementUnread()
	}
	b.incrementWptr()
	b.incrementSize()
	if r == 0 {
		// Park the write pointer so read sees it hit EOF.
		b.wptr = -1
	}
	return nil
}

func (b *RuneBuffer) shouldPushRptr() bool {
	return b.unread > 0 && b.unread == b.size && b.wptr == b.rptr
}

func (b *RuneBuffer) incrementRptr() {
	b.rptr = b.normalizePtr(b.rptr + 1)
}

func (b *RuneBuffer) decrementRptr() {
	b.rptr = b.normalizePtr(b.rptr - 1)
}

func (b *RuneBuffer) incrementWptr() {
	b.wptr = b.normalizePtr(b.wptr + 1)
}

func (b *RuneBuffer) incrementUnread() {
	b.unread++
}

func (b *RuneBuffer) decrementUnread() {
	b.unread--
}

func (b *RuneBuffer) incrementSize() {
	if b.size == len(b.buf) {
		return
	}
	b.size++
}

func (b *RuneBuffer) normalizePtr(p int) int {
	bufSize := len(b.buf)
	for p < 0 {
		p += bufSize
	}
	for p >= bufSize {
		p -= bufSize
	}
	return p
}

// ThreadSafeRuneBuffer just puts a sync.Mutex in front of public operations of a RuneBuffer.
type ThreadSafeRuneBuffer struct {
	*RuneBuffer

	mux sync.Mutex
}

func (t *ThreadSafeRuneBuffer) ReadRune() (rune, error) {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.RuneBuffer.ReadRune()
}

func (t *ThreadSafeRuneBuffer) UnreadRune() {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.RuneBuffer.UnreadRune()
}

func (t *ThreadSafeRuneBuffer) UnreadNumRunes(num int) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.RuneBuffer.UnreadNumRunes(num)
}
