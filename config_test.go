package lunash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAllConfigs(t *testing.T) {
	configs, err := LoadAllConfigs(exampleConfigPath)
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(configs))

		assert.Equal(t, "hsm1", configs[0].Nickname)
		assert.Equal(t, "1.1.1.1", configs[0].Hostname)
		assert.Equal(t, 22, configs[0].SSHport)
		assert.Equal(t, "admin", configs[0].SSHlogin)
		assert.Equal(t, "password", configs[0].SSHpassword)
		assert.Equal(t, "SHA256:40QvIN7FAGgYxl+5UVoTskPK1zKswZcDzPCK6aZReuU", configs[0].SSHfingerprint)
		assert.Equal(t, "other_password", configs[0].Password)

		assert.Equal(t, "", configs[1].Nickname)
		assert.Equal(t, "2.2.2.2", configs[1].Hostname)
		assert.Equal(t, 2222, configs[1].SSHport)
		assert.Equal(t, "somebody", configs[1].SSHlogin)
		assert.Equal(t, "s3cret", configs[1].SSHpassword)
		assert.Equal(t, "SHA256:yv0u4ILh4aYY1GHGQpXu025WbZgUpJ0FBhjw18SgZPE", configs[1].SSHfingerprint)
		assert.Equal(t, "other_s3cret", configs[1].Password)
	}

	_, err = LoadAllConfigs("./doesnt_exist.json")
	assert.NotNil(t, err)

	_, err = LoadAllConfigs("./glide.yaml")
	assert.NotNil(t, err)
}

func TestLoadConfigs(t *testing.T) {
	names := []string{"hsm1", "2.2.2.2"}
	configs, err := LoadConfigs(exampleConfigPath, names)
	if assert.Nil(t, err) {
		assert.Equal(t, 2, len(configs))
		assert.Equal(t, "hsm1", configs[0].Nickname)
		assert.Equal(t, "2.2.2.2", configs[1].Hostname)
	}

	names = []string{"hsm1"}
	configs, err = LoadConfigs(exampleConfigPath, names)
	if assert.Nil(t, err) {
		assert.Equal(t, 1, len(configs))
		assert.Equal(t, "hsm1", configs[0].Nickname)
	}

	names = []string{"2.2.2.2"}
	configs, err = LoadConfigs(exampleConfigPath, names)
	if assert.Nil(t, err) {
		assert.Equal(t, 1, len(configs))
		assert.Equal(t, "2.2.2.2", configs[0].Hostname)
	}

	_, err = LoadConfigs("./doesnt_exist.json", names)
	assert.NotNil(t, err)

	_, err = LoadConfigs("./glide.yaml", names)
	assert.NotNil(t, err)
}
