package lunash

import (
	"bytes"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const lunashPrompt = "lunash:>"

func readUntilPrompt(stdout io.Reader) (string, error) {
	buf := bytes.NewBuffer(nil)
	chunk := make([]byte, 1024)

	for {
		n, err := stdout.Read(chunk)
		if err != nil {
			return "", errors.Wrap(err, "Error reading from stdout")
		}
		buf.Write(chunk[:n])

		if strings.HasSuffix(buf.String(), lunashPrompt) {
			break
		}
	}

	// Strip the prompt.
	_, withoutPrompt := lastLine(buf.String())

	return withoutPrompt, nil
}

func lastLine(str string) (lline, rest string) {
	lines := strings.Split(str, "\n")
	lline = strings.TrimSuffix(lines[len(lines)-1], "\r")
	rest = strings.Join(lines[:len(lines)-1], "\n")
	return
}

func firstLine(str string) (fline, rest string) {
	lines := strings.Split(str, "\n")
	fline = strings.TrimSuffix(lines[0], "\r")
	rest = strings.Join(lines[1:], "\n")
	return
}
