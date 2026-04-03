//go:build darwin

package platform

import (
	"github.com/openwith-org/openwith/internal/defaults"
)

// DarwinBackend wraps the existing defaults.Client for macOS.
type DarwinBackend struct {
	client *defaults.Client
}

func NewDarwinBackend(client *defaults.Client) *DarwinBackend {
	return &DarwinBackend{client: client}
}

func (b *DarwinBackend) GetDefault(ext string) string {
	defs := b.client.GetAllDefaults([]string{ext})
	if len(defs) > 0 {
		return defs[0].AppName
	}
	return "unknown"
}

func (b *DarwinBackend) GetDefaultBundleID(ext string) string {
	raw := b.client.GetAllDefaultsRaw([]string{ext})
	return raw[ext]
}

func (b *DarwinBackend) SetDefault(ext string, appID string) error {
	changes := map[string]string{ext: appID}
	results := make(chan defaults.Result)
	go b.client.ApplyChanges(changes, results)
	for r := range results {
		if !r.Success {
			return r.Error
		}
	}
	return nil
}

func (b *DarwinBackend) GetDefaultBrowser() string {
	return b.client.GetDefaultBrowser()
}

func (b *DarwinBackend) SetDefaultBrowser(appID string) error {
	return b.client.SetDefaultBrowser(appID)
}

func (b *DarwinBackend) DetectApp(appID string) bool {
	return defaults.ValidateBundleID(appID)
}
