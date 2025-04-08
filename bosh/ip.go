package bosh

import (
	"net/netip"
)

type IP struct {
	ip netip.Addr
}

func ParseIP(ip string) (IP, error) {
	parsed, err := netip.ParseAddr(ip)
	if err != nil {
		return IP{}, err
	}

	return IP{
		ip: parsed,
	}, nil
}

func (i IP) Add(offset int) IP {
	next := i.ip
	for range offset {
		next = next.Next()
	}

	return IP{
		ip: next,
	}
}

func (i IP) Subtract(offset int) IP {
	prev := i.ip
	for range offset {
		prev = prev.Prev()
	}

	return IP{
		ip: prev,
	}
}

func (i IP) String() string {
	return i.ip.String()
}
