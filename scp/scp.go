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

// PutFile writes a file on the remote server.
func PutFile(session *ssh.Session, path string, file []byte) error {
	debugf("PutFile: %s\n", path)

	stdin, stdout, err := openPipes(session)
	if err != nil {
		return err
	}
	defer stdin.Close()

	cmd := "scp -t " + path
	debugf("Running command: '%s'\n", cmd)
	if err = session.Start(cmd); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error starting command '%s'", cmd))
	}

	return putFile(stdin, stdout, path, file)
}

func putFile(stdin io.WriteCloser, stdout io.Reader, path string, file []byte) error {
	debug("Reading bytes")
	buf := make([]byte, 1024)
	n, err := stdout.Read(buf)
	if err != nil {
		return errors.Wrap(err, "Error reading bytes from stdout")
	}
	debug("Read bytes: ", buf[:n])

	// Eg. "C0644 1192 server.pem"
	hdr := fmt.Sprintf("C0644 %d %s\n", len(file), path)
	debugf("Writing file header to stdin: %s", strconv.QuoteToASCII(hdr))
	if _, err = stdin.Write([]byte(hdr)); err != nil {
		return errors.Wrap(err, "Error writing metadata to stdin")
	}
	debug("Wrote header")

	debug("Writing file")
	if _, err = stdin.Write(file); err != nil {
		return errors.Wrap(err, "Error writing file to stdin")
	}
	debug("Wrote file")

	debug("Reading bytes")
	n, err = stdout.Read(buf)
	if err != nil {
		return errors.Wrap(err, "Error reading bytes from stdout")
	}
	debug("Read bytes: ", buf[:n])

	debug("Writing reply")
	if _, err = stdin.Write([]byte{0x00}); err != nil {
		return errors.Wrap(err, "Error writing reply to stdin")
	}
	debug("Wrote reply")

	debug("Reading bytes")
	n, err = stdout.Read(buf)
	if err != nil {
		return errors.Wrap(err, "Error reading bytes from stdout")
	}
	debug("Read bytes: ", buf[:n])

	return nil
}

// GetFile get's a file from the remote server.
func GetFile(session *ssh.Session, path string) ([]byte, error) {
	debugf("GetFile: %s\n", path)

	stdin, stdout, err := openPipes(session)
	if err != nil {
		return nil, err
	}
	defer stdin.Close()

	cmd := "scp -f " + path
	debugf("Running command: '%s'\n", cmd)
	if err = session.Start(cmd); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error starting command '%s'", cmd))
	}

	return getFile(stdin, stdout)
}

func getFile(stdin io.WriteCloser, stdout io.Reader) ([]byte, error) {
	debug("Writing null byte")
	stdin.Write([]byte{0x00})

	debug("Reading file header")
	hdr := make([]byte, 1024)
	n, err := stdout.Read(hdr)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading header from stdin")
	}
	debugf("Got header: '%s'", strconv.QuoteToASCII(string(hdr[:n])))

	// Parse file length from header
	// Eg.
	//   C0644 1192 server.pem
	//   mode| size| path
	hparts := strings.SplitN(string(hdr[:n]), " ", 3)
	if hparts == nil || len(hparts) != 3 {
		return nil, errors.New("Bad header from SCP")
	}

	flen, err := strconv.ParseInt(hparts[1], 10, 64)
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
