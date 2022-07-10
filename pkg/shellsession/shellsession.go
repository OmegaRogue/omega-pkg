package shellsession

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"io"
	"os/exec"
)

type Session struct {
	Cmd *exec.Cmd

	Stdin  bufio.Reader
	Stdout bufio.Writer
	Stderr bufio.Writer
}

type PipeReadWriteCloser struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

func (p2 *PipeReadWriteCloser) Read(p []byte) (int, error) {

	n, err := p2.reader.Read(p)
	if err != nil {
		return n, errors.Wrap(err, "error on read")
	}
	return n, nil
}

func (p2 *PipeReadWriteCloser) Write(p []byte) (int, error) {
	n, err := p2.writer.Write(p)
	if err != nil {
		return n, errors.Wrap(err, "error on write")
	}
	return n, nil
}

func (p2 *PipeReadWriteCloser) Close() error {
	if err := p2.writer.Close(); err != nil {
		return errors.Wrap(err, "error on close writer")
	}
	if err := p2.reader.Close(); err != nil {
		return errors.Wrap(err, "error on close reader")
	}
	return nil
}

func NewPipeReadWriter() *PipeReadWriteCloser {
	r, w := io.Pipe()
	rw := new(PipeReadWriteCloser)
	rw.writer = w
	rw.reader = r
	return rw
}

func Start() {

	exec.CommandContext(context.TODO(), "/bin/sh", "-c")

}
