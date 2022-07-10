package muxio

import (
	"github.com/pkg/errors"
	"io"
	"sync"
)

type MuxedWriter struct {
	writers []io.Writer
	mu      sync.Mutex
	i       int
}

func (m *MuxedWriter) RemuxI(i int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i >= len(m.writers) || i < 0 {
		return errors.New("mux index out of range")
	}
	m.i = i
	return nil
}
func (m *MuxedWriter) Remux(writer io.Writer) error {
	return m.RemuxI(m.GetWriterMux(writer))
}

func (m *MuxedWriter) GetMux() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.i
}
func (m *MuxedWriter) AddWriter(writer io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writers = append(m.writers, writer)
}
func (m *MuxedWriter) GetWriterMux(writer io.Writer) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.writers {
		if writer == v {
			return k
		}
	}
	return -1
}
func (m *MuxedWriter) RemoveWriterI(i int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writers = append(m.writers[:i], m.writers[i+1:]...)
}
func (m *MuxedWriter) RemoveWriter(writer io.Writer) {
	m.RemoveWriterI(m.GetWriterMux(writer))
}

func (m *MuxedWriter) Write(p []byte) (n int, err error) {
	i := m.GetMux()
	return m.writers[i].Write(p)
}

type MuxedReader struct {
	readers []io.Reader
	mu      sync.Mutex
	i       int
}

func (m *MuxedReader) RemuxI(i int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i >= len(m.readers) || i < 0 {
		return errors.New("mux index out of range")
	}
	m.i = i
	return nil
}
func (m *MuxedReader) Remux(reader io.Reader) error {
	return m.RemuxI(m.GetReaderMux(reader))
}

func (m *MuxedReader) GetMux() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.i
}
func (m *MuxedReader) AddReader(reader io.Reader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readers = append(m.readers, reader)
}
func (m *MuxedReader) GetReaderMux(reader io.Reader) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.readers {
		if reader == v {
			return k
		}
	}
	return -1
}
func (m *MuxedReader) RemoveReaderI(i int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readers = append(m.readers[:i], m.readers[i+1:]...)
}
func (m *MuxedReader) RemoveReader(reader io.Reader) {
	m.RemoveReaderI(m.GetReaderMux(reader))
}

func (m *MuxedReader) Read(p []byte) (n int, err error) {
	i := m.GetMux()
	return m.readers[i].Read(p)
}
