package fuse

import (
	"embedpi/config"

	"github.com/wmarbut/go-epdfuse"
	"go.uber.org/zap"
)

type FuseDevice struct {
	EpdFuse epdfuse.EpdFuse
	config.FuseCfg
}

// NewDevice creates a new fuse device
func NewDevice(cfg config.FuseCfg) (f FuseDevice) {
	f.Active = cfg.Active
	if f.Active {
		zap.S().Info("Papirus is enable !")
		f.FuseCfg = cfg
		f.EpdFuse = epdfuse.NewCustomEpdFuse(cfg.Device, cfg.Width, cfg.Height)
	} else {
		zap.S().Info("Papirus is disable !")
	}

	return f
}

// WriteData on fuse device with given data
func (f FuseDevice) WriteData(details string) error {
	f.EpdFuse.Clear()
	return f.EpdFuse.WriteText(details)
}
