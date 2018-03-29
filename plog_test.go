package plog

import (
	"sync"
	"testing"
)

// TestLogger tests Append usage
// This function does not test Print, PrintDef, Println, or Printf because they are all
// one or two lines that call Buffer functions
func TestLogger(t *testing.T) {
	t.Run("concurrent append", testLoggerConcurrentAppend)
}

func testLoggerConcurrentAppend(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	l := NewLogger(rb)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			l.Lock()
			defer l.Unlock()
			l.Append("n")
			l.Append("e")
			l.Append("m")
			l.Append("o")
			l.AppendDone(Major)
		}(&wg)
	}
	wg.Wait()

	popWithExpected("nemo", rb, t)
	popWithExpected("nemo", rb, t)
	popWithExpected("nemo", rb, t)
}

// TestRingBuffer runs a variety of subtests covering RingBuffer usage
func TestRingBuffer(t *testing.T) {
	t.Run("get priority", testRingBufferGetPriority)
	t.Run("set priority", testRingBufferSetPriority)
	t.Run("write", testRingBufferWrite)
	t.Run("pwrite", testRingBufferPWrite)
	t.Run("pop", testRingBufferPop)
	t.Run("overflow", testRingBufferOverflow)
}

func testRingBufferGetPriority(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	p := rb.GetPriority()
	if p != Minor {
		t.Logf("expected %v, got %v\n", Minor, p)
		t.Fail()
	}
}

func testRingBufferSetPriority(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	rb.SetPriority(Major)
	p := rb.GetPriority()
	if p != Major {
		t.Logf("expected %v, got %v\n", Major, p)
		t.Fail()
	}
}

func testRingBufferWrite(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	rb.Write([]byte("nemo"))
	popWithExpected("nemo", rb, t)
}

func testRingBufferPWrite(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	rb.PWrite(Major, []byte("nemo"))
	popWithExpected("nemo", rb, t)
}

// testRingBufferPop inserts some strings in random order and asserts that they are
// popped in the correct order
func testRingBufferPop(t *testing.T) {
	rb := NewRingBuffer(Minor, 3)
	rb.PWrite(Major, []byte("major0"))
	rb.Write([]byte("minor0"))
	rb.PWrite(Trivial, []byte("trivial0"))
	rb.PWrite(Critical, []byte("critical0"))
	rb.PWrite(Major, []byte("major1"))
	rb.Write([]byte("minor1"))
	rb.PWrite(Critical, []byte("critical1"))
	rb.PWrite(Trivial, []byte("trivial1"))

	popWithExpected("critical1", rb, t)
	popWithExpected("critical0", rb, t)
	popWithExpected("major1", rb, t)
	popWithExpected("major0", rb, t)
	popWithExpected("minor1", rb, t)
	popWithExpected("minor0", rb, t)
	popWithExpected("trivial1", rb, t)
	popWithExpected("trivial0", rb, t)
}

// testRingBufferOverflow asserts that RingBuffer displays proper overflow behavior
// for a ring buffer that has had more entries than its capacity inserted
func testRingBufferOverflow(t *testing.T) {
	rb := NewRingBuffer(Minor, 5)
	rb.Write([]byte("0"))
	rb.Write([]byte("1"))
	rb.Write([]byte("2"))
	rb.Write([]byte("3"))
	rb.Write([]byte("4"))
	rb.Write([]byte("5"))
	rb.Write([]byte("6"))

	popWithExpected("6", rb, t)
	popWithExpected("5", rb, t)
	popWithExpected("4", rb, t)
	popWithExpected("3", rb, t)
	popWithExpected("2", rb, t)
	if _, err := rb.Pop(); err == nil {
		t.Log("err should not be nil")
		t.Fail()
	}
}

// popWithExpected is a quick helper method for making the above test code easier to read
func popWithExpected(expected string, rb *RingBuffer, t *testing.T) {
	if s, err := rb.Pop(); err != nil || s != expected {
		t.Logf("err: %v || %s != %s\n", err, s, expected)
		t.Fail()
	}
}
