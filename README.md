## Priority Logger ##

This package 'plog' implements a Logger struct that can be printed to in much the same as as log.Logger.  Rather than printing to an io.Writer, it prints to a plog.Buffer that supports writing with priority information and does not implement io.Reader.  Rather, plog.Buffer has a Pop() function that returns the highest priority and latest log printed to it (in that order) as a string.

For the most part, this package has some odd features as they were pulled from a Java coding challenge.