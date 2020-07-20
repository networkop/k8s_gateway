package gateway

import (
	"context"

	"strconv"

	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin(thisPlugin)

const thisPlugin = "k8s_gateway"

func init() {
	plugin.Register(thisPlugin, setup)
}

func setup(c *caddy.Controller) error {
	gw, err := parse(c)
	if err != nil {
		return plugin.Error(thisPlugin, err)
	}

	gw.Controller, err = RunKubeController(context.Background())
	if err != nil {
		return plugin.Error(thisPlugin, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		gw.Next = next
		return gw
	})

	return nil
}

func parse(c *caddy.Controller) (*Gateway, error) {
	gw := newGateway()

	for c.Next() {
		zones := c.RemainingArgs()
		gw.Zones = zones

		if len(gw.Zones) == 0 {
			gw.Zones = make([]string, len(c.ServerBlockKeys))
			copy(gw.Zones, c.ServerBlockKeys)
		}

		for i, str := range gw.Zones {
			gw.Zones[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "resources":
				args := c.RemainingArgs()

				gw.updateResources(args)

				if len(args) == 0 {
					return nil, c.Errf("Incorrectly formated 'resource' parameter")
				}
			case "ttl":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				t, err := strconv.Atoi(args[0])
				if err != nil {
					return nil, err
				}
				if t < 0 || t > 3600 {
					return nil, c.Errf("ttl must be in range [0, 3600]: %d", t)
				}
				gw.ttl = uint32(t)
			default:
				return nil, c.Errf("Unknown property '%s'", c.Val())
			}
		}
	}
	return gw, nil

}
