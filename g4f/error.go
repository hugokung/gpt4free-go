package g4f

import (
	"errors"
	"io"
)

var (
	ErrStreamRestart = errors.New("stream restart")
	StreamCompleted  = io.EOF
)
