package lunash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/errors"
)

// Config stores information about a single HSM configuration.
type Config struct {
	Nickname       string `json:"nickname"`
	Hostname       string `json:"hostname"`
	SSHport        int    `json:"ssh_port"`
	SSHlogin       string `json:"ssh_login"`
	SSHpassword    string `json:"ssh_password"`
	SSHfingerprint string `json:"ssh_fingerprint"`
	Password       string `json:"hsm_password"`
}

// LoadAllConfigs loads all Configs from a config file.
func LoadAllConfigs(path string) ([]*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening config file")
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading config file")
	}

	var configs []*Config
	if err = json.Unmarshal(buf.Bytes(), &configs); err != nil {
		return nil, errors.Wrap(err, "Error parsing config file")
	}

	return configs, nil
}

// LoadConfigs loads the configs for the HSMs with the given nicknames or
// hostnames.
func LoadConfigs(path string, names []string) ([]*Config, error) {
	all, err := LoadAllConfigs(path)
	if err != nil {
		return nil, err
	}

	configs := make([]*Config, 0, len(all))

	for _, config := range all {
		for _, name := range names {
			if config.Hostname == name || config.Nickname == name {
				configs = append(configs, config)
			}
		}
	}

	return configs, nil
}

// Client returns a Client from this config.
func (cfg *Config) Client() *Client {
	return newClient(cfg)
}

func (cfg *Config) sshClient() (*ssh.Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Hostname, cfg.SSHport)

	ccfg := &ssh.ClientConfig{
		User:            cfg.SSHlogin,
		HostKeyCallback: cfg.verifyPublicKey,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.SSHpassword),
		},
	}

	client, err := ssh.Dial("tcp", address, ccfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening SSH connection to "+cfg.Hostname)
	}

	return client, nil
}

// verifyPublicKey is a callback for verifying the HSM's SSH public key.
func (cfg *Config) verifyPublicKey(host string, _ net.Addr, key ssh.PublicKey) error {
	actual := ssh.FingerprintSHA256(key)

	if actual != cfg.SSHfingerprint {
		return fmt.Errorf("Bad HSM SSH public key. Host: %s Fingerprint: %s", host, actual)
	}

	return nil
}
