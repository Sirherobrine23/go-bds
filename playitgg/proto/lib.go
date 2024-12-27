package proto

import (
	"encoding/binary"
	"errors"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/playitgg/message_encoding"
)

type AgentSessionId struct {
	SessionID, AccountID, AgentID uint64
}

func (sesion AgentSessionId) MarshalBinary() ([]byte, error) {
	return slices.Concat(binary.BigEndian.AppendUint64(nil, sesion.SessionID), binary.BigEndian.AppendUint64(nil, sesion.AccountID), binary.BigEndian.AppendUint64(nil, sesion.AgentID)), nil
}
func (sesion *AgentSessionId) Read(data []byte) (int, error) {
	sesion.SessionID, sesion.AccountID, sesion.AgentID = binary.BigEndian.Uint64(data[:8]), binary.BigEndian.Uint64(data[8:16]), binary.BigEndian.Uint64(data[16:24])
	return 24, nil
}

type PortProto int

const (
	PortProtoBoth PortProto = iota
	PortProtoTcp
	PortProtoUdp
)

func (port PortProto) MarshalBinary() ([]byte, error) {
	switch port {
	case PortProtoTcp:
		return []byte{1}, nil
	case PortProtoUdp:
		return []byte{2}, nil
	case PortProtoBoth:
		return []byte{3}, nil
	default:
		return []byte{}, nil
	}
}

func (port *PortProto) Read(data []byte) (int, error) {
	switch data[0] {
	case 1:
		*port = PortProtoTcp
	case 2:
		*port = PortProtoUdp
	case 3:
		*port = PortProtoBoth
	default:
		return 1, errors.New("invalid port proto")
	}
	return 1, nil
}

type PortRange struct {
	IP                 message_encoding.Addr
	PortStart, PortEnd uint16
	Proto              PortProto
}

func (portRange PortRange) MarshalBinary() ([]byte, error) {
	ipBin, err := portRange.IP.MarshalBinary()
	if err != nil {
		return nil, err
	}
	protoc, err := portRange.Proto.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return slices.Concat(
		ipBin,
		binary.BigEndian.AppendUint16(nil, portRange.PortStart),
		binary.BigEndian.AppendUint16(nil, portRange.PortEnd),
		protoc,
	), nil
}

func (portRange *PortRange) Read(data []byte) (n int, _ error) {
	{
		n1, err := portRange.IP.Read(data)
		data = data[:n1]
		n += n1
		if err != nil {
			return n, err
		}
	}
	portRange.PortStart, portRange.PortEnd = binary.BigEndian.Uint16(data[0:2]), binary.BigEndian.Uint16(data[2:4])
	n += 4
	data = data[:4]
	{
		n1, err := portRange.Proto.Read(data)
		n += n1
		if err != nil {
			return n, err
		}
	}
	return
}
