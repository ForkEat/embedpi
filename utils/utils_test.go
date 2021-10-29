package utils

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestGenerateSSID(t *testing.T) {
	ssid := GenerateSSID()
	assert.Regexp(t, "^fork-eat|[0-9]{8}$", ssid)
}

func TestGeneratePassword(t *testing.T) {
	password := GeneratePassword()
	assert.Regexp(t, "^[a-zA-Z0-9]{8}$", password)
}
