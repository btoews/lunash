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
	defer stdin.Close()

	cmd := "scp -f " + path
	debugf("Running command: '%s'\n", cmd)
	if err = session.Start(cmd); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error starting command '%s'", cmd))
	}
	defer session.Wait()

	return getFile(stdin, stdout)
}

func getFile(stdin io.WriteCloser, stdout io.Reader) ([]byte, error) {
	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	debug("Reading file metadata")
	meta := make([]byte, 1024)
	n, err := stdout.Read(meta)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading from stdin")
	}
	debugf("Got metadata: '%s'", strconv.QuoteToASCII(string(meta[:n])))

	// Parse file length from metadata
	// Eg.
	//   C0644 1192 server.pem
	//   mode| size| path
	mparts := strings.SplitN(string(meta[:n]), " ", 3)
	if mparts == nil || len(mparts) != 3 {
		return nil, errors.New("Bad metadata from SCP")
	}

	flen, err := strconv.ParseInt(mparts[1], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing SCP metadata")
	}

	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	debug("Reading file")
	file := make([]byte, flen)
	n, err = stdout.Read(file)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading from stdin")
	}
	if n != int(flen) {
		return nil, errors.New("Read incomplete file")
	}
	debug("File: ", string(file))

	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	return file, nil
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
