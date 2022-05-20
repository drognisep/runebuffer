package runebuffer

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestScenario(t *testing.T) {
	reader := strings.NewReader("ABCD")
	ring := NewRuneBufferWithSize(reader, 3)

	t.Run("Can't unread with an empty buffer", func(t *testing.T) {
		ring.UnreadRune()
		assert.Equal(t, 0, ring.rptr)
		assert.Equal(t, 0, ring.wptr)
	})

	t.Run("Reading the first rune in the stream sets index 0 of the buffer", func(t *testing.T) {
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'A', read)
		assert.Equal(t, 1, ring.rptr)
		assert.Equal(t, 1, ring.wptr)
	})

	t.Run("Can't unread beyond buffer size", func(t *testing.T) {
		ring.UnreadNumRunes(100)
		assert.Equal(t, 0, ring.rptr)
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'A', read)
		assert.Equal(t, 1, ring.rptr)
		assert.Equal(t, 1, ring.wptr)
	})

	t.Run("Reading the second rune continues the readStream + read", func(t *testing.T) {
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'B', read)
		assert.Equal(t, 2, ring.rptr)
		assert.Equal(t, 2, ring.wptr)
	})

	t.Run("Unread twice and read 3", func(t *testing.T) {
		ring.UnreadNumRunes(2)
		assert.Equal(t, 0, ring.rptr)
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'A', read)
		read, err = ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'B', read)
		read, err = ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'C', read)
		assert.Equal(t, 0, ring.rptr)
		assert.Equal(t, 0, ring.wptr)
	})

	t.Run("Unread 3 times and call the underlying readStream to ensure it pushes rptr", func(t *testing.T) {
		ring.UnreadNumRunes(3)
		assert.Equal(t, 0, ring.rptr)
		assert.NoError(t, ring.readStream())
		assert.Equal(t, 1, ring.rptr)
		assert.Equal(t, 1, ring.wptr)
	})

	t.Run("Reading to EOF will return 0 and nil error", func(t *testing.T) {
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'B', read)
		read, err = ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'C', read)
		read, err = ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, 'D', read)
		read, err = ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, rune(0), read)
		assert.Equal(t, 2, ring.rptr)
		assert.Equal(t, -1, ring.wptr)
	})

	t.Run("Subsequent reads from buffer will continue to return 0 and nil error", func(t *testing.T) {
		read, err := ring.ReadRune()
		assert.NoError(t, err)
		assert.Equal(t, rune(0), read)
		assert.Equal(t, 2, ring.rptr)
		assert.Equal(t, -1, ring.wptr)
	})
}
