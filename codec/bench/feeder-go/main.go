package main

import (
	"bufio"
	"bytes"
	"os"

	"github.com/streamingfast/cli"
)

func main() {
	data := os.Getenv("PAYLOAD")
	cli.Ensure(data != "", "PAYLOAD environment variable must be set")

	content, err := os.ReadFile(data)
	cli.NoError(err, "unable to read payload file %q", data)

	lines := bytes.Split(content, []byte("\n"))

	bufferedStdout := bufio.NewWriterSize(os.Stdout, 64*1024)
	defer bufferedStdout.Flush()

	for {
		for _, line := range lines {
			bufferedStdout.Write(line)
			bufferedStdout.WriteByte('\n')
		}
	}
}
