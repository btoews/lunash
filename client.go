package lunash

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Client is an SSH session with the HSM.
type Client struct {
	config *Config
	client *ssh.Client
}

// newClient creates a Session from a Config.
func newClient(config *Config) *Client {
	return &Client{config: config}
}

// Connect connects to the HSM.
func (c *Client) Connect() (err error) {
	c.client, err = c.config.sshClient()
	return
}

// Run runs multiple commands in an SSH PTY session and returns their outputs.
func (c *Client) Run(commands []string, login bool) ([]string, error) {
	var runErr, ptyErr error
	outputs := make([]string, 0, len(commands))

	ptyErr = c.WithPTY(func(stdin io.WriteCloser, stdout io.Reader) {
		if _, err := readUntilPrompt(stdout); err != nil {
			runErr = errors.Wrap(err, "Error reading shell banner")
			return
		}

		if login {
			_, err := stdin.Write([]byte(fmt.Sprintf("hsm login -p %s\n", c.config.Password)))
			if err != nil {
				runErr = errors.Wrap(err, "Error running 'hsm login'")
				return
			}

			_, err = readUntilPrompt(stdout)
			if err != nil {
				runErr = errors.New("Error reading command output for 'hsm login'")
				return
			}
		}

		for _, cmd := range commands {
			var cmdWithNewline string

			if strings.HasSuffix(cmd, "\n") {
				cmdWithNewline = cmd
			} else {
				cmdWithNewline = cmd + "\n"
			}

			_, err := stdin.Write([]byte(cmdWithNewline))
			if err != nil {
				runErr = errors.Wrap(err, fmt.Sprintf("Error sending command '%s'", cmd))
				return
			}

			output, err := readUntilPrompt(stdout)
			if err != nil {
				runErr = fmt.Errorf("Error reading command output for '%s'", cmd)
				return
			}

			// strip the command itself from the output.
			_, output = firstLine(output)

			// Check return code.
			status, output := lastLine(output)
			const success = "Command Result : 0 (Success)"
			if status != success {
				runErr = fmt.Errorf("Non-success return code while running '%s'", cmd)
			}

			outputs = append(outputs, output)
		}
	})

	if runErr != nil {
		return outputs, runErr
	}
	return outputs, ptyErr
}

// WithPTY calls the callback with an PTY SSH session.
func (c *Client) WithPTY(cb func(io.WriteCloser, io.Reader)) error {
	var ptyErr, sesErr error

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	sesErr = c.WithSession(func(session *ssh.Session) {
		// Request pseudo terminal
		if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
			ptyErr = errors.Wrap(err, "Error requesting PTY")
			return
		}

		stdin, err := session.StdinPipe()
		if err != nil {
			ptyErr = errors.Wrap(err, "Error getting stdin pipe")
			return
		}

		stdout, err := session.StdoutPipe()
		if err != nil {
			ptyErr = errors.Wrap(err, "Error getting stdout pipe")
			return
		}

		if err = session.Shell(); err != nil {
			ptyErr = errors.Wrap(err, "Error starting login shell")
			return
		}

		cb(stdin, stdout)
	})

	if ptyErr != nil {
		return ptyErr
	}
	return sesErr
}

// WithSession calls the callback with an SSH session.
func (c *Client) WithSession(cb func(*ssh.Session)) error {
	session, err := c.client.NewSession()
	if err != nil {
		return errors.Wrap(err, "Error opening session")
	}

	cb(session)

	err = session.Close()
	if err != nil {
		return errors.Wrap(err, "Error closing session")
	}

	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return errors.Wrap(err, "Error closing client")
		}
	}

	return nil
}
