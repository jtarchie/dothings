package writers

import (
	"bytes"
	"io"
	"sync"
)

type stringReaderRouter struct {
	sync.Mutex
	buffer *bytes.Buffer
	router io.Writer
}

func NewStringReaderRouter(router io.Writer) *stringReaderRouter {
	b := bytes.NewBufferString("")
	return &stringReaderRouter{
		buffer: b,
		router: io.MultiWriter(router, b),
	}
}

func (b *stringReaderRouter) String() string {
	b.Lock()
	defer b.Unlock()

	return b.buffer.String()
}

func (b *stringReaderRouter) Write(p []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()

	return b.router.Write(p)
}
