package bosh

import (
	"encoding/binary"
	"math"
	"net/netip"
)

type CIDRBlock struct {
	cidr netip.Prefix
}

func ParseCIDRBlock(cidrBlock string) (CIDRBlock, error) {
	prefix, err := netip.ParsePrefix(cidrBlock)
	if err != nil {
		return CIDRBlock{}, err
	}
	return CIDRBlock{cidr: prefix}, nil
}

func (c CIDRBlock) GetFirstIP() IP {
	return c.GetNthIP(0)
}

func (c CIDRBlock) GetNthIP(n int) IP {
	ip := IP{c.cidr.Addr()}

	if n > 0 {
		return ip.Add(n)
	}
	return ip.Subtract(n)

}

func (c CIDRBlock) GetLastIP() IP {
	a := c.cidr.Addr()
	if a.Is4() {
		four := a.As4()
		uint32Four := binary.BigEndian.Uint32(four[:])
		masklen := c.cidr.Addr().BitLen() - c.cidr.Bits()
		mask := uint32(math.Pow(2, float64(masklen))) - 1
		uint32Four += mask
		binary.BigEndian.PutUint32(four[:], uint32Four)
		return IP{netip.AddrFrom4(four)}
	}
	panic("not implemented")
}
