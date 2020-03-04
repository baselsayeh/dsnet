package dsnet

import (
	"net"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func Up() {
	conf := MustLoadDsnetConfig()
	CreateLink(conf)
	ConfigureDevice(conf)
}

func CreateLink(conf *DsnetConfig) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = conf.InterfaceName

	link := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}

	err := netlink.LinkAdd(link)
	if err != nil {
		ExitFail("Could not add interface '%s' (%v)", conf.InterfaceName, err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   conf.IP,
			Mask: conf.Network.IPNet.Mask,
		},
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		ExitFail("Could not add addr %s to interface %s", addr.IP, err)
	}

	// bring up interface (UNKNOWN state instead of UP, a wireguard quirk)
	err = netlink.LinkSetUp(link)
}

func ConfigureDevice(conf *DsnetConfig) {
	wgConfig := wgtypes.Config{
		PrivateKey:   &conf.PrivateKey.Key,
		ListenPort:   &conf.ListenPort,
		ReplacePeers: true,
		Peers:        conf.GetWgPeerConfigs(),
	}

	wg, err := wgctrl.New()
	check(err)
	defer wg.Close()

	err = wg.ConfigureDevice(conf.InterfaceName, wgConfig)

	if err != nil {
		ExitFail("Could not configure device '%s' (%v)", conf.InterfaceName, err)
	}

	if err != nil {
		ExitFail("Could not bring up device '%s' (%v)", conf.InterfaceName, err)
	}
}
