package g4f

import (
	"errors"
	"io"
)

var (
	ErrModelType     = errors.New("model type doesn't exist")
	ErrStreamRestart = errors.New("stream restart")
	StreamEOF        = io.EOF
	StreamCompleted  = errors.New("stream completed")
	ErrResponse      = errors.New("response error")
)
