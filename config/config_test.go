package config

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/caarlos0/env"
)

func TestIsDev(t *testing.T) {
	assert.True(t, IsDev("dev"))
	assert.True(t, IsDev("development"))
	assert.False(t, IsDev("prod"))
	assert.False(t, IsDev("production"))
	assert.False(t, IsDev("anything else"))
}

func TestDefaultAppConfig(t *testing.T) {
	appConfig := AppConfig{}
	err := env.Parse(&appConfig)

	assert.NoError(t, err)
	assert.Equal(t, appConfig, AppConfig{Port: "8080", CfgUrl: "config.json"})
}

func TestAppConfigIsDev(t *testing.T) {
	c := AppConfig{
		Environment: "dev",
	}
	assert.True(t, c.IsDev())

	c.Environment = "development"
	assert.True(t, c.IsDev())

	c.Environment = "prod"
	assert.False(t, c.IsDev())

	c.Environment = "production"
	assert.False(t, c.IsDev())

	c.Environment = "anything else"
	assert.False(t, c.IsDev())
}

func TestLoad_fileExist(t *testing.T) {
	_, err := LoadCfg("../config.json")
	assert.NoError(t, err)
}

func TestLoad_fileNotExist(t *testing.T) {
	_, err := LoadCfg("ghk.json")
	assert.Error(t, err)
}

func TestLoad2(t *testing.T) {
	appConfig, err := getDefaultConfig()

	assert.NoError(t, err)
	assert.Equal(t, appConfig, &SetupCfg{
		WpaSupplicantCfg: WpaSupplicantCfg{
			CfgFile: "/etc/wpa_supplicant/wpa_supplicant.conf",
		},
		HostApdCfg: HostApdCfg{
			Ip:      "192.168.27.1",
			Channel: "6",
		},
		DnsmasqCfg: DnsmasqCfg{
			Address:     "/#/192.168.27.1",
			DhcpRange:   "192.168.27.100,192.168.27.150,1h",
			VendorClass: "set:device,IoT",
		},
	})
}

func getDefaultConfig() (*SetupCfg, error) {
	v := &SetupCfg{}
	err := json.Unmarshal([]byte(`
		{
			"environment": "dev",
			"dnsmasq_cfg": {
				"address": "/#/192.168.27.1",
				"dhcp_range": "192.168.27.100,192.168.27.150,1h",
				"vendor_class": "set:device,IoT"
			},
			"host_apd_cfg": {
				"ip": "192.168.27.1",
				"channel": "6"
			},
			"wpa_supplicant_cfg": {
				"cfg_file": "/etc/wpa_supplicant/wpa_supplicant.conf"
			}
		}`), v)
	return v, err
}
