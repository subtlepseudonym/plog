package plog

import (
	"io"
	"log"
)

type LoggerT struct {
	*log.Logger
}

const (
	Lpriority = 1 << 7 // FIXME: this will totally break if log adds more flags
)

func New(out io.Writer, prefix string, flag int) *LoggerT {
	return &LoggerT{log.New(out, prefix, flag)}
}
