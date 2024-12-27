package message_encoding

import (
	"encoding"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"
	"slices"
)

type Binary interface {
	io.Reader
	encoding.BinaryMarshaler
}

func MarshalBinaryAppend(buff []byte, input encoding.BinaryMarshaler) ([]byte, error) {
	n1, err := input.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	return slices.Concat(buff, n1), nil
}

var _ Binary = &AddrPort{}

type AddrPort netip.AddrPort

func (addr AddrPort) Addr() netip.Addr              { return netip.AddrPort(addr).Addr() }
func (addr AddrPort) AppendTo(b []byte) []byte      { return netip.AddrPort(addr).AppendTo(b) }
func (addr AddrPort) Compare(p2 netip.AddrPort) int { return netip.AddrPort(addr).Compare(p2) }
func (addr AddrPort) IsValid() bool                 { return netip.AddrPort(addr).IsValid() }
func (addr AddrPort) Port() uint16                  { return netip.AddrPort(addr).Port() }
func (addr AddrPort) String() string                { return netip.AddrPort(addr).String() }
func (addr AddrPort) MarshalText() ([]byte, error)  { return netip.AddrPort(addr).MarshalText() }
func (addr *AddrPort) UnmarshalText(text []byte) error {
	return (*netip.AddrPort)(addr).UnmarshalText(text)
}

func (addr AddrPort) MarshalBinary() ([]byte, error) {
	buff := []byte{0}
	if netip.AddrPort(addr).Addr().Is4() {
		buff[0] = 4
	} else {
		buff[0] = 6
	}
	return binary.BigEndian.AppendUint16(netip.AddrPort(addr).Addr().AppendTo(buff), netip.AddrPort(addr).Port()), nil
}

func (addr *AddrPort) Read(data []byte) (n int, err error) {
	var add Addr
	{
		n1, err := add.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n1, err
		}
	}
	n += 2
	*addr = AddrPort(netip.AddrPortFrom(netip.Addr(add), binary.BigEndian.Uint16(data[:2]))) // Add to struct
	return
}

type Addr netip.Addr

func (addr Addr) MarshalBinary() ([]byte, error) {
	buff := []byte{0}
	if netip.Addr(addr).Is6() {
		buff[0] = 6
	} else {
		buff[0] = 4
	}
	return netip.Addr(addr).AppendTo(buff), nil
}

func (addr *Addr) Read(data []byte) (int, error) {
	switch data[0] {
	case 4:
		*addr = Addr(netip.AddrFrom4([4]byte(data[1:5])))
		return 5, nil
	case 6:
		*addr = Addr(netip.AddrFrom16([16]byte(data[1:17])))
		return 17, nil
	default:
		return 1, errors.New("invalid Address")
	}
}

func (ip Addr) AppendTo(b []byte) []byte           { return netip.Addr(ip).AppendTo(b) }
func (ip Addr) As16() (a16 [16]byte)               { return netip.Addr(ip).As16() }
func (ip Addr) As4() (a4 [4]byte)                  { return netip.Addr(ip).As4() }
func (ip Addr) AsSlice() []byte                    { return netip.Addr(ip).AsSlice() }
func (ip Addr) BitLen() int                        { return netip.Addr(ip).BitLen() }
func (ip Addr) Compare(ip2 netip.Addr) int         { return netip.Addr(ip).Compare(ip2) }
func (ip Addr) Is4() bool                          { return netip.Addr(ip).Is4() }
func (ip Addr) Is4In6() bool                       { return netip.Addr(ip).Is4In6() }
func (ip Addr) Is6() bool                          { return netip.Addr(ip).Is6() }
func (ip Addr) IsGlobalUnicast() bool              { return netip.Addr(ip).IsGlobalUnicast() }
func (ip Addr) IsInterfaceLocalMulticast() bool    { return netip.Addr(ip).IsInterfaceLocalMulticast() }
func (ip Addr) IsLinkLocalMulticast() bool         { return netip.Addr(ip).IsLinkLocalMulticast() }
func (ip Addr) IsLinkLocalUnicast() bool           { return netip.Addr(ip).IsLinkLocalUnicast() }
func (ip Addr) IsLoopback() bool                   { return netip.Addr(ip).IsLoopback() }
func (ip Addr) IsMulticast() bool                  { return netip.Addr(ip).IsMulticast() }
func (ip Addr) IsPrivate() bool                    { return netip.Addr(ip).IsPrivate() }
func (ip Addr) IsUnspecified() bool                { return netip.Addr(ip).IsUnspecified() }
func (ip Addr) IsValid() bool                      { return netip.Addr(ip).IsValid() }
func (ip Addr) Less(ip2 netip.Addr) bool           { return netip.Addr(ip).Less(ip2) }
func (ip Addr) MarshalText() ([]byte, error)       { return netip.Addr(ip).MarshalText() }
func (ip Addr) Next() netip.Addr                   { return netip.Addr(ip).Next() }
func (ip Addr) Prefix(b int) (netip.Prefix, error) { return netip.Addr(ip).Prefix(b) }
func (ip Addr) Prev() netip.Addr                   { return netip.Addr(ip).Prev() }
func (ip Addr) String() string                     { return netip.Addr(ip).String() }
func (ip Addr) StringExpanded() string             { return netip.Addr(ip).StringExpanded() }
func (ip Addr) Unmap() netip.Addr                  { return netip.Addr(ip).Unmap() }
func (ip *Addr) UnmarshalText(text []byte) error   { return (*netip.Addr)(ip).UnmarshalText(text) }
func (ip Addr) WithZone(zone string) netip.Addr    { return netip.Addr(ip).WithZone(zone) }
func (ip Addr) Zone() string                       { return netip.Addr(ip).Zone() }

type Buffer []byte

func (buff Buffer) MarshalBinary() ([]byte, error) {
	return slices.Concat(binary.BigEndian.AppendUint64(nil, uint64(len(buff))), buff), nil
}

func (buff *Buffer) Read(data []byte) (n int, err error) {
	size := int(binary.BigEndian.Uint64(data[:8]))
	*buff = make(Buffer, size)
	copy(*buff, data)
	return size + 8, nil
}
