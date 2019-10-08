package writers

import (
	"io"
	"strings"
	"sync"

	"github.com/jtarchie/dothings/executor"
)

type stringWriter struct {
	sync.Mutex
	*strings.Builder
}

func (s *stringWriter) String() string {
	s.Lock()
	defer s.Unlock()

	return s.Builder.String()
}

func (s *stringWriter) Write(p []byte) (int, error) {
	s.Lock()
	defer s.Unlock()

	return s.Builder.Write(p)
}

type inMemory struct {
	sync.Mutex
	writers map[string]*stringWriter
}

var _ executor.Writer = &inMemory{}

func NewInMemory() *inMemory {
	return &inMemory{
		writers: make(map[string]*stringWriter),
	}
}

func (h *inMemory) GetWriter(task executor.Tasker) (io.Writer, io.Writer) {
	h.Lock()
	defer h.Unlock()

	if writer, ok := h.writers[task.ID()]; ok {
		return writer, writer
	}

	writer := &stringWriter{
		Builder: &strings.Builder{},
	}
	h.writers[task.ID()] = writer
	return writer, writer
}

func (h *inMemory) GetString(task executor.Tasker) (string, string) {
	h.Lock()
	defer h.Unlock()

	if writer, ok := h.writers[task.ID()]; ok {
		output := writer.String()
		return output, output
	}

	return "", ""
}
