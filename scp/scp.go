package scp

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Debug controls whether we log verbosely
var Debug bool

// GetFile get's a file from the remote server.
func GetFile(session *ssh.Session, path string) ([]byte, error) {
	debugf("GetFile: %s\n", path)

	stdin, stdout, err := openPipes(session)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting file with SCP")
	}

	output := make(chan []byte)
	gerr := make(chan error)

	go getFile(stdin, stdout, output, gerr)

	cmd := "scp -f " + path
	debugf("Running command: '%s'\n", cmd)
	if err = session.Run(cmd); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error running command '%s'", cmd))
	}
	debug("Done")

	select {
	case o := <-output:
		return o, nil
	case err := <-gerr:
		return nil, err
	}
}

func getFile(stdin io.WriteCloser, stdout io.Reader, output chan []byte, gerr chan error) {
	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	debug("Reading file metadata")
	meta := make([]byte, 1024)
	n, err := stdout.Read(meta)
	if err != nil {
		gerr <- errors.Wrap(err, "Error reading from stdin")
		return
	}
	debugf("Got metadata: '%s'", strconv.QuoteToASCII(string(meta[:n])))

	// Parse file length from metadata
	// Eg.
	//   C0644 1192 server.pem
	//   mode| size| path
	mparts := strings.SplitN(string(meta[:n]), " ", 3)
	if mparts == nil || len(mparts) != 3 {
		gerr <- errors.New("Bad metadata from SCP")
		return
	}

	flen, err := strconv.ParseInt(mparts[1], 10, 64)
	if err != nil {
		gerr <- errors.Wrap(err, "Error parsing SCP metadata")
		return
	}

	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	debug("Reading file")
	file := make([]byte, flen)
	n, err = stdout.Read(file)
	if err != nil {
		gerr <- errors.Wrap(err, "Error reading from stdin")
		return
	}
	if n != int(flen) {
		gerr <- errors.New("Read incomplete file")
		return
	}
	debug("File: ", string(file))

	debug("Closing stdin")
	stdin.Close()

	output <- file
}

func openPipes(session *ssh.Session) (stdin io.WriteCloser, stdout io.Reader, err error) {
	stdin, err = session.StdinPipe()
	if err != nil {
		err = errors.Wrap(err, "Error opening stdin pipe")
		return
	}

	stdout, err = session.StdoutPipe()
	if err != nil {
		err = errors.Wrap(err, "Error opening stdin pipe")
		return
	}

	return
}

func debug(v ...interface{}) {
	if Debug {
		log.Println(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}
