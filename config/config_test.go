package config

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert"
)

func TestIsDev(t *testing.T) {
	assert.True(t, IsDev("dev"))
	assert.True(t, IsDev("development"))
	assert.False(t, IsDev("prod"))
	assert.False(t, IsDev("production"))
	assert.False(t, IsDev("anything else"))
}

func TestAppConfigIsDev(t *testing.T) {
	c := SetupCfg{
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
		Environment: "dev",
		FuseCfg: FuseCfg{
			Active: true,
			Device: "/dev/epd",
			Width:  264,
			Height: 176,
		},
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
			"fuse_cfg": {
				"active": true,
				"device": "/dev/epd",
				"width": 264,
				"height": 176
			},
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
