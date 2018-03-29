package plog

import (
	"bytes"
	"container/ring"
	"fmt"
	"sync"
)

// LogPriority is a simple enum for determining the order in which Logger releases
// logs from the buffer
type LogPriority int

const (
	Trivial LogPriority = iota
	Minor
	Major
	Critical
)

// Logger stores logs in buffer interface and enables writing to that buffer
type Logger struct {
	buf  Buffer
	aBuf *bytes.Buffer
	lock *sync.Mutex
}

// NewLogger returns a reference to a newly allocated Logger struct
func NewLogger(b Buffer) *Logger {
	return &Logger{
		buf:  b,
		aBuf: bytes.NewBuffer([]byte{}),
		lock: &sync.Mutex{},
	}
}

// Lock exposes the Logger's internal mutex Lock() function
func (l *Logger) Lock() {
	l.lock.Lock()
}

// Unlock exposes the Logger's internal mutex Unlock() function
func (l *Logger) Unlock() {
	l.lock.Unlock()
}

// Append appends a string to Logger's append buffer
// It's a good idea to call l.Lock() before entering this function
func (l *Logger) Append(s string) {
	l.aBuf.WriteString(s)
}

// AppendDone signals that the caller is done appending to the current ring buffer
// value and that the ring buffer reference should be updated.
// The Logger's Lock() function should be called prior to using this function
func (l *Logger) AppendDone(p LogPriority) {
	l.buf.PWrite(p, l.aBuf.Bytes())
	l.aBuf.Reset()
}

// Print inserts s into the p priority ring buffer and updates the Logger's reference
// to the ring buffer
func (l *Logger) Print(p LogPriority, s string) {
	l.buf.PWrite(p, []byte(s))
}

// PrintDef operates the same way as Logger.Print, but uses the Buffer's set Priority
func (l *Logger) PrintDef(s string) {
	l.Print(l.buf.GetPriority(), s)
}

// Println addends a newline to s and calls l.Print
func (l *Logger) Println(p LogPriority, s string) {
	l.Print(p, s+"\n")
	// not sure if using '+' to concat in this case is that much worse than the overhead
	// required for writing / copying to a buffer (or maybe appending to a byte slice?)
}

// Printf applies formatting to format before passing it to l.Print
func (l *Logger) Printf(p LogPriority, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Print(p, s)
}

// GetBuffer returns the reference to the Logger's internal Buffer
func (l *Logger) GetBuffer() Buffer {
	return l.buf
}

// Buffer allows you to define custom write and output behavior while still implementing
// the io.Writer interface for use with other packages
type Buffer interface {
	Pop() (string, error)
	Write([]byte) (int, error)
	PWrite(LogPriority, []byte) (int, error)
	GetPriority() LogPriority
	SetPriority(LogPriority)
}

// RingBuffer uses a ring buffer to store logs and manage memory usage
// This buffer optimizes for write performance over read performance
type RingBuffer struct {
	p      LogPriority
	bufCap int
	buf    map[int]*ring.Ring
	lock   *sync.Mutex
	highP  int // current highest priority value
}

// NewRingBuffer initializes a new RingBuffer struct with the given LogPriority and
// buffer size and returns a reference to it
func NewRingBuffer(p LogPriority, size int) *RingBuffer {
	return &RingBuffer{
		p:      p,
		bufCap: size,
		buf:    make(map[int]*ring.Ring),
		lock:   &sync.Mutex{},
		highP:  0,
	}
}

// GetPriority returns the RingBuffer's LogPriority
func (r *RingBuffer) GetPriority() LogPriority {
	return r.p
}

// SetPriority sets the RingBuffer's default priority
func (r *RingBuffer) SetPriority(p LogPriority) {
	r.p = p
}

// Pop returns the RingBuffer's contents prioritizing higher priority and newer
// logs first
func (r *RingBuffer) Pop() (string, error) {
	_, ok := r.buf[r.highP]
	if !ok {
		return "", fmt.Errorf("Buffer is empty")
	}

	r.buf[r.highP] = r.buf[r.highP].Prev()
	b, ok := r.buf[r.highP].Value.([]byte)
	if !ok {
		return "", fmt.Errorf("pop type assertion failed")
	}

	r.buf[r.highP].Value = nil

	// update highP
	for i := r.highP; i >= 0; i-- {
		if _, ok := r.buf[i]; ok && r.buf[i].Prev().Value != nil {
			r.highP = i
			break
		}
	}

	return string(b), nil
}

// Write write a slice of bytes (p) into it's ring buffer
func (r *RingBuffer) Write(b []byte) (int, error) {
	return r.PWrite(r.GetPriority(), b)
}

// PWrite writes to the ring buffer with priority p
func (r *RingBuffer) PWrite(p LogPriority, b []byte) (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	i := int(p)
	if i > r.highP {
		r.highP = i
	}

	if r.buf[i] == nil {
		r.buf[i] = ring.New(r.bufCap)
	}

	r.buf[i].Value = b
	r.buf[i] = r.buf[i].Next()

	return len(b), nil
}
