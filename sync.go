package dsnet

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Sync() {
	// TODO check device settings first
	conf := MustLoadDsnetConfig()
	MustConfigureDevice(conf)
}

// MustConfigureDevice is like ConfigureDevice, except it exits on
// error
func MustConfigureDevice(conf *DsnetConfig) {
	err := ConfigureDevice(conf)
	check(err)
}

// ConfigureDevice sets up the WG interface
func ConfigureDevice(conf *DsnetConfig) error {
	wg, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer wg.Close()

	dev, err := wg.Device(conf.InterfaceName)

	if err != nil {
		return fmt.Errorf("Could not retrieve device '%s' (%v)", conf.InterfaceName, err)
	}

	peers := conf.GetWgPeerConfigs()

	// compare peers to see if any exist on the device and not the config. If
	// so, they should be removed by appending a dummy peer with Remove:true + pubkey.
	knownKeys := make(map[wgtypes.Key]bool)

	for _, peer := range peers {
		knownKeys[peer.PublicKey] = true
	}

	// find deleted peers, and append dummy "remove" peers
	for _, peer := range dev.Peers {
		if !knownKeys[peer.PublicKey] {
			peers = append(peers, wgtypes.PeerConfig{
				PublicKey: peer.PublicKey,
				Remove:    true,
			})
		}
	}

	wgConfig := wgtypes.Config{
		PrivateKey: &conf.PrivateKey.Key,
		ListenPort: &conf.ListenPort,
		// ReplacePeers with the same peers results in those peers losing
		// connection, so it's not possible to do declarative configuration
		// idempotently with ReplacePeers like I had assumed. Instead, peers
		// must be removed imperatively with Remove:true. Peers can still be
		// added/updated with ConfigureDevice declaratively.
		ReplacePeers: false,
		Peers:        peers,
	}

	err = wg.ConfigureDevice(conf.InterfaceName, wgConfig)

	if err != nil {
		return fmt.Errorf("Could not configure device '%s' (%v)", conf.InterfaceName, err)
	}
	return nil
}
