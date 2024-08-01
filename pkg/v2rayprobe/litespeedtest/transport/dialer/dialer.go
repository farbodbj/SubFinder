package dialer

import (
	"context"
	"errors"
	"net"

	"github.com/xxf098/lite-proxy/transport/resolver"
)

func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	case "tcp4", "tcp6", "udp4", "udp6":
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		dialer, err := Dialer()
		if err != nil {
			return nil, err
		}

		var ip net.IP
		switch network {
		case "tcp4", "udp4":
			ip, err = resolver.ResolveIPv4(host)
		default:
			ip, err = resolver.ResolveIPv6(host)
		}

		if err != nil {
			return nil, err
		}

		return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	case "tcp", "udp":
		return dualStackDialContext(ctx, network, address)
	default:
		return nil, errors.New("network invalid")
	}
}

func ListenPacket(network, address string) (net.PacketConn, error) {
	// cfg := &net.ListenConfig{}
	// return cfg.ListenPacket(context.Background(), network, address)
	return effectiveListener.ListenPacket(context.Background(), network, address)
}

func dualStackDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	returned := make(chan struct{})
	defer close(returned)

	type dialResult struct {
		net.Conn
		error
		resolved bool
		ipv6     bool
		done     bool
	}
	results := make(chan dialResult)
	var primary, fallback dialResult

	startRacer := func(ctx context.Context, network, host string, ipv6 bool) {
		result := dialResult{ipv6: ipv6, done: true}
		defer func() {
			select {
			case results <- result:
			case <-returned:
				if result.Conn != nil {
					result.Conn.Close()
				}
			}
		}()

		dialer, err := Dialer()
		if err != nil {
			result.error = err
			return
		}

		var ip net.IP
		if ipv6 {
			ip, result.error = resolver.ResolveIPv6(host)
		} else {
			ip, result.error = resolver.ResolveIPv4(host)
		}
		if result.error != nil {
			return
		}
		result.resolved = true

		result.Conn, result.error = dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	go startRacer(ctx, network+"4", host, false)
	go startRacer(ctx, network+"6", host, true)

	for res := range results {
		if res.error == nil {
			return res.Conn, nil
		}

		if !res.ipv6 {
			primary = res
		} else {
			fallback = res
		}

		if primary.done && fallback.done {
			if primary.resolved {
				return nil, primary.error
			} else if fallback.resolved {
				return nil, fallback.error
			} else {
				return nil, primary.error
			}
		}
	}

	return nil, errors.New("never touched")
}
