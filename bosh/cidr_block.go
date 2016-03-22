package bosh

import (
	"fmt"
	"strconv"
	"strings"
)

type CIDRBlock struct {
	CIDRSize int
	firstIP  IP
}

func ParseCIDRBlock(cidrBlock string) (CIDRBlock, error) {
	const HIGHEST_BITMASK = 32
	const CIDR_PARTS = 2

	cidrParts := strings.Split(cidrBlock, "/")

	if len(cidrParts) != CIDR_PARTS {
		return CIDRBlock{}, fmt.Errorf(`"%s" cannot parse CIDR block`, cidrBlock)
	}

	ip, err := ParseIP(cidrParts[0])
	if err != nil {
		return CIDRBlock{}, err
	}

	maskBits, err := strconv.Atoi(cidrParts[1])
	if err != nil {
		return CIDRBlock{}, err
	}

	if maskBits < 0 || maskBits > HIGHEST_BITMASK {
		return CIDRBlock{}, fmt.Errorf("mask bits out of range")
	}

	cidrSize := 1 << (HIGHEST_BITMASK - uint(maskBits))
	return CIDRBlock{
		CIDRSize: cidrSize,
		firstIP:  ip,
	}, nil
}

func (c CIDRBlock) GetFirstIP() IP {
	return c.firstIP
}

func (c CIDRBlock) GetLastIP() IP {
	return c.firstIP.Add(c.CIDRSize - 1)
}
