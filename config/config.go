package config

import (
	"embed/utils"
	"encoding/json"
	"io/ioutil"
)

type AppConfig struct {
	Port   string `default:"8080" env:"PORT"`
	CfgUrl string `default:"config.json" env:"CFG_URL"`
}

// FuseCfg is the eink config file.
type FuseCfg struct {
	Active bool   `json:"active"`
	Device string `json:"device"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SetupCfg is the main configuration structure.
type SetupCfg struct {
	DnsmasqCfg       DnsmasqCfg       `json:"dnsmasq_cfg"`
	HostApdCfg       HostApdCfg       `json:"host_apd_cfg"`
	WpaSupplicantCfg WpaSupplicantCfg `json:"wpa_supplicant_cfg"`
	FuseCfg          FuseCfg          `json:"fuse_cfg"`
	Environment      string           `json:"environment"`
}

// DnsmasqCfg configures dnsmasq and is used by SetupCfg.
type DnsmasqCfg struct {
	Address     string `json:"address"`      // --address=/#/192.168.27.1",
	DhcpRange   string `json:"dhcp_range"`   // "--dhcp-range=192.168.27.100,192.168.27.150,1h",
	VendorClass string `json:"vendor_class"` // "--dhcp-vendorclass=set:device,IoT",
}

// HostApdCfg configures hostapd and is used by SetupCfg.
type HostApdCfg struct {
	Ssid          string
	WpaPassphrase string
	Channel       string `json:"channel"` //  channel=6
	Ip            string `json:"ip"`      // 192.168.27.1
}

// WpaSupplicantCfg configures wpa_supplicant and is used by SetupCfg
type WpaSupplicantCfg struct {
	CfgFile string `json:"cfg_file"` // /etc/wpa_supplicant/wpa_supplicant.conf
}

func (s *SetupCfg) IsDev() bool {
	return IsDev(s.Environment)
}

//IsDev return true if application is on dev stack
func IsDev(env string) bool {
	return env == "dev" || env == "development"
}

// LoadCfg loads the config file.
func LoadCfg(cfgLocation string) (*SetupCfg, error) {

	v := &SetupCfg{}
	var jsonData []byte

	fileData, err := ioutil.ReadFile(cfgLocation)
	if err != nil {
		return nil, err
	}

	jsonData = fileData

	err = json.Unmarshal(jsonData, v)

	v.HostApdCfg.WpaPassphrase = utils.GeneratePassword()
	v.HostApdCfg.Ssid = utils.GenerateSSID()

	return v, err
}
