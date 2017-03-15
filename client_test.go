package lunash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	configs, _ := LoadConfigs(exampleConfigPath, []string{"hsm1"})
	client := newClient(configs[0])
	assert.Equal(t, configs[0], client.config)
}
