package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/playitgg/message_encoding"
)

var (
	_ message_encoding.Binary = &ControlResponse{}
	_ message_encoding.Binary = &Pong{}
	_ message_encoding.Binary = &AgentRegistered{}
	_ message_encoding.Binary = &AgentPortMapping{}
	_ message_encoding.Binary = &UdpChannelDetails{}
)

type ControlResponse struct {
	Pong                                                         *Pong
	InvalidSignature, Unauthorized, RequestQueued, TryAgainLater bool
	AgentRegistered                                              *AgentRegistered
	AgentPortMapping                                             *AgentPortMapping
	UdpChannelDetails                                            *UdpChannelDetails
}

func (response ControlResponse) MarshalBinary() ([]byte, error) {
	data := []byte{}
	switch {
	case response.Pong != nil:
		data = binary.BigEndian.AppendUint32(nil, 1)
	case response.InvalidSignature:
		data = binary.BigEndian.AppendUint32(nil, 2)
	case response.Unauthorized:
		data = binary.BigEndian.AppendUint32(nil, 3)
	case response.RequestQueued:
		data = binary.BigEndian.AppendUint32(nil, 4)
	case response.TryAgainLater:
		data = binary.BigEndian.AppendUint32(nil, 5)
	case response.AgentRegistered != nil:
		data = binary.BigEndian.AppendUint32(nil, 6)
	case response.AgentPortMapping != nil:
		data = binary.BigEndian.AppendUint32(nil, 7)
	case response.UdpChannelDetails != nil:
		data = binary.BigEndian.AppendUint32(nil, 8)
	}

	return data, nil
}

func (response *ControlResponse) Read(data []byte) (n int, _ error) {
	n = 4
	switch binary.BigEndian.Uint32(data[:4]) {
	case 1:
		response.Pong = &Pong{}
		n2, err := response.Pong.Read(data[4:])
		if n += n2; err != nil {
			return n, err
		}
	case 2:
		response.InvalidSignature = true
	case 3:
		response.Unauthorized = true
	case 4:
		response.RequestQueued = true
	case 5:
		response.TryAgainLater = true
	case 6:
		response.AgentRegistered = &AgentRegistered{}
		n2, err := response.AgentRegistered.Read(data[4:])
		if n += n2; err != nil {
			return n, err
		}
	case 7:
		response.AgentPortMapping = &AgentPortMapping{}
		n2, err := response.AgentPortMapping.Read(data[4:])
		if n += n2; err != nil {
			return n, err
		}
	case 8:
		response.UdpChannelDetails = &UdpChannelDetails{}
		n2, err := response.UdpChannelDetails.Read(data[4:])
		if n += n2; err != nil {
			return n, err
		}
	}
	return
}

type Pong struct {
	RequestNow      uint64
	ServerNow       uint64
	ServerID        uint64
	DataCenterID    uint32
	ClientAddr      message_encoding.AddrPort
	TunnelAddr      message_encoding.AddrPort
	SessionExpireAt *uint64
}

func (pong Pong) MarshalBinary() ([]byte, error) {
	n2, err := pong.ClientAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	n3, err := pong.TunnelAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}

	data := slices.Concat(
		binary.BigEndian.AppendUint64(nil, pong.RequestNow),
		binary.BigEndian.AppendUint64(nil, pong.ServerNow),
		binary.BigEndian.AppendUint64(nil, pong.ServerID),
		binary.BigEndian.AppendUint32(nil, pong.DataCenterID),
		n2,
		n3,
	)

	if pong.SessionExpireAt == nil {
		data = append(data, 0)
	} else {
		data = append(data, 1)
		data = slices.Concat(data, binary.BigEndian.AppendUint64(nil, *pong.SessionExpireAt))
	}

	return data, nil
}

func (pong *Pong) Read(data []byte) (n int, _ error) {
	pong.RequestNow = binary.BigEndian.Uint64(data[:8])
	pong.ServerNow = binary.BigEndian.Uint64(data[8:16])
	pong.ServerID = binary.BigEndian.Uint64(data[16:24])
	pong.DataCenterID = binary.BigEndian.Uint32(data[24:28])
	data = data[28:]
	n = 28

	{
		pong.ClientAddr = message_encoding.AddrPort{}
		n1, err := pong.ClientAddr.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n, err
		}
	}
	{
		n1, err := pong.ClientAddr.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n, err
		}
	}
	{
		n += 1
		switch data[0] {
		case 1:
			pong.SessionExpireAt = new(uint64)
			*pong.SessionExpireAt = binary.BigEndian.Uint64(data[:8])
			n += 8
		}
	}

	return
}

type AgentRegistered struct {
	ID       AgentSessionId
	ExpireAt uint64
}

func (agent AgentRegistered) MarshalBinary() ([]byte, error) {
	data, err := agent.ID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append(data, binary.BigEndian.AppendUint64(nil, agent.ExpireAt)...), nil
}

func (agent *AgentRegistered) Read(data []byte) (n int, _ error) {
	{
		n1, err := agent.ID.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n, err
		}
	}
	agent.ExpireAt = binary.BigEndian.Uint64(data[0:8])
	n += 8
	return
}

type AgentPortMappingFound struct {
	ToAgent *AgentSessionId
}

func (found AgentPortMappingFound) MarshalBinary() ([]byte, error) {
	if found.ToAgent != nil {
		m2, err := found.ToAgent.MarshalBinary()
		if m2 == nil {
			return append([]byte{1}, m2...), err
		}
		return nil, err
	}
	return []byte{}, nil
}

func (found *AgentPortMappingFound) Read(data []byte) (n int, _ error) {
	n = 1
	switch data[0] {
	case 1:
		found.ToAgent = &AgentSessionId{}
		{
			n1, err := found.ToAgent.Read(data[1:])
			n += n1
			return n, err
		}
	default:
		return 1, errors.New("unknown AgentPortMappingFound id")
	}
}

type AgentPortMapping struct {
	Range PortRange
	Found *AgentPortMappingFound
}

func (agent AgentPortMapping) MarshalBinary() ([]byte, error) {
	rangeData, err := agent.Range.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if agent.Found == nil {
		rangeData = append(rangeData, byte(0))
	} else {
		rangeData = append(rangeData, byte(1))
		data, err := agent.Found.MarshalBinary()
		if err != nil {
			return nil, err
		}
		rangeData = append(rangeData, data...)
	}

	return rangeData, nil
}
func (agent *AgentPortMapping) Read(data []byte) (n int, _ error) {
	agent.Range = PortRange{}
	{
		n1, err := agent.Range.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n, err
		}
	}
	switch data[0] {
	case 0:
		n++
		return
	case 1:
		agent.Found = &AgentPortMappingFound{}
		n1, err := agent.Found.Read(data[1:])
		n += n1
		if err != nil {
			return n, err
		}
	}
	return
}

type UdpChannelDetails struct {
	TunnelAddr message_encoding.AddrPort
	Token      message_encoding.Buffer
}

func (udp UdpChannelDetails) MarshalBinary() ([]byte, error) {
	sockAddr, err := udp.TunnelAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	token, _ := udp.Token.MarshalBinary()
	return append(sockAddr, token...), nil
}

func (udp *UdpChannelDetails) Read(data []byte) (n int, _ error) {
	udp.TunnelAddr = message_encoding.AddrPort{}
	{
		n1, err := udp.TunnelAddr.Read(data)
		data = data[:n1]
		n += n1
		if err != nil {
			return n, err
		}
	}
	udp.Token = []byte{}
	{
		n1, err := udp.Token.Read(data)
		n += n1
		if err != nil {
			return n, err
		}
	}
	return
}

type ControlRequestId int

const (
	ControlRequestId_PingV1 ControlRequestId = iota + 1
	ControlRequestIdAgentRegisterV1
	ControlRequestIdAgentKeepAliveV1
	ControlRequestIdSetupUdpChannelV1
	ControlRequestIdAgentCheckPortMappingV1
	ControlRequestIdPingV2
	ControlRequestIdEND
)

func (rid ControlRequestId) MarshalBinary() ([]byte, error) {
	switch rid {
	case ControlRequestId_PingV1:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestId_PingV1)), nil
	case ControlRequestIdAgentRegisterV1:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdAgentRegisterV1)), nil
	case ControlRequestIdAgentKeepAliveV1:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdAgentKeepAliveV1)), nil
	case ControlRequestIdSetupUdpChannelV1:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdSetupUdpChannelV1)), nil
	case ControlRequestIdAgentCheckPortMappingV1:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdAgentCheckPortMappingV1)), nil
	case ControlRequestIdPingV2:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdPingV2)), nil
	case ControlRequestIdEND:
		return binary.BigEndian.AppendUint32(nil, uint32(ControlRequestIdEND)), nil
	}
	return []byte{}, nil
}

func (rid *ControlRequestId) Read(data []byte) (n int, _ error) {
	switch ControlRequestId(binary.BigEndian.Uint32(data[0:4])) {
	case ControlRequestId_PingV1:
		*rid = ControlRequestId_PingV1
	case ControlRequestIdAgentRegisterV1:
		*rid = ControlRequestIdAgentRegisterV1
	case ControlRequestIdAgentKeepAliveV1:
		*rid = ControlRequestIdAgentKeepAliveV1
	case ControlRequestIdSetupUdpChannelV1:
		*rid = ControlRequestIdSetupUdpChannelV1
	case ControlRequestIdAgentCheckPortMappingV1:
		*rid = ControlRequestIdAgentCheckPortMappingV1
	case ControlRequestIdPingV2:
		*rid = ControlRequestIdPingV2
	case ControlRequestIdEND:
		*rid = ControlRequestIdEND
	}
	return 4, nil
}

type AgentCheckPortMapping struct {
	AgentSessionID AgentSessionId
	PortRange      PortRange
}

func (portCheck AgentCheckPortMapping) MarshalBinary() ([]byte, error) {
	data, err := portCheck.AgentSessionID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ndata, err := portCheck.PortRange.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return slices.Concat(data, ndata), nil
}

func (portCheck *AgentCheckPortMapping) Read(data []byte) (n int, _ error) {
	portCheck.AgentSessionID = AgentSessionId{}
	n1, err := portCheck.AgentSessionID.Read(data)
	data = data[:n1]
	if n += n1; err != nil {
		return n, err
	}
	portCheck.PortRange = PortRange{}
	n1, err = portCheck.PortRange.Read(data)
	if n += n1; err != nil {
		return n, err
	}
	return
}

type Ping struct {
	Now         uint64
	CurrentPing *uint64
	SessionID   *AgentSessionId
}

func (ping Ping) MarshalBinary() ([]byte, error) {
	data := binary.BigEndian.AppendUint64(nil, ping.Now)
	if ping.CurrentPing == nil {
		data = append(data, byte(0))
	} else {
		data = append(data, byte(1))
		data = binary.BigEndian.AppendUint64(data, *ping.CurrentPing)
	}
	if ping.SessionID == nil {
		data = append(data, byte(0))
	} else {
		data = append(data, byte(1))
		m1, err := ping.SessionID.MarshalBinary()
		if err != nil {
			return nil, err
		}
		data = append(data, m1...)
	}
	return data, nil
}

func (ping *Ping) Read(data []byte) (n int, _ error) {
	n = 8
	ping.Now = binary.BigEndian.Uint64(data[:8])
	n++
	if data[9] == 1 {
		*ping.CurrentPing = binary.BigEndian.Uint64(data[9:17])
		n += 8
	}
	n++
	if data[18] == 1 {
		ping.SessionID = &AgentSessionId{}
		n1, err := ping.SessionID.Read(data[18:])
		if n += n1; err != nil {
			return n, err
		}
	}
	return
}

type AgentRegister struct {
	AccountID, AgentID, AgentVersion, Timestamp uint64
	ClientAddr, TunnelAddr                      message_encoding.AddrPort
	Signature                                   [32]byte
}

func (agent AgentRegister) MarshalBinary() (data []byte, err error) {
	data = binary.BigEndian.AppendUint64(nil, agent.AccountID)
	data = binary.BigEndian.AppendUint64(data, agent.AgentID)
	data = binary.BigEndian.AppendUint64(data, agent.AgentVersion)
	data = binary.BigEndian.AppendUint64(data, agent.Timestamp)
	if data, err = message_encoding.MarshalBinaryAppend(data, agent.ClientAddr); err != nil {
		return data, err
	} else if data, err = message_encoding.MarshalBinaryAppend(data, agent.TunnelAddr); err != nil {
		return data, err
	} else if bytes.Equal(agent.Signature[:], make([]byte, 32)) {
		return data, errors.New("failed to write full signature")
	}
	data = slices.Concat(data, agent.Signature[:])
	return
}

func (agent *AgentRegister) Read(data []byte) (n int, _ error) {
	agent.AccountID = binary.BigEndian.Uint64(data[:8])
	agent.AgentID = binary.BigEndian.Uint64(data[8:16])
	agent.AgentVersion = binary.BigEndian.Uint64(data[16:24])
	agent.Timestamp = binary.BigEndian.Uint64(data[24:32])
	n += 32
	data = data[32:]
	{
		agent.ClientAddr = message_encoding.AddrPort{}
		n1, err := agent.ClientAddr.Read(data)
		n += n1
		if data = data[n1:]; err != nil {
			return n, err
		}
	}
	{
		agent.TunnelAddr = message_encoding.AddrPort{}
		n1, err := agent.TunnelAddr.Read(data)
		n += n1
		if data = data[n1:]; err != nil {
			return n, err
		}
	}
	agent.Signature = [32]byte(data[:32])
	n += 32
	return
}

type ControlRequest struct {
	Ping                  *Ping
	AgentRegister         *AgentRegister
	AgentKeepAlive        *AgentSessionId
	SetupUdpChannel       *AgentSessionId
	AgentCheckPortMapping *AgentCheckPortMapping
}

func (request ControlRequest) MarshalBinary() (data []byte, err error) {
	switch {
	case request.Ping != nil:
		if data, err = ControlRequestIdPingV2.MarshalBinary(); err != nil {
			return
		}
		data, err = message_encoding.MarshalBinaryAppend(data, request.Ping)
	case request.AgentRegister != nil:
		if data, err = ControlRequestIdAgentRegisterV1.MarshalBinary(); err != nil {
			return
		}
		data, err = message_encoding.MarshalBinaryAppend(data, request.AgentRegister)
	case request.AgentKeepAlive != nil:
		if data, err = ControlRequestIdAgentKeepAliveV1.MarshalBinary(); err != nil {
			return
		}
		data, err = message_encoding.MarshalBinaryAppend(data, request.AgentKeepAlive)
	case request.SetupUdpChannel != nil:
		if data, err = ControlRequestIdSetupUdpChannelV1.MarshalBinary(); err != nil {
			return
		}
		data, err = message_encoding.MarshalBinaryAppend(data, request.SetupUdpChannel)
	case request.AgentCheckPortMapping != nil:
		if data, err = ControlRequestIdAgentCheckPortMappingV1.MarshalBinary(); err != nil {
			return
		}
		data, err = message_encoding.MarshalBinaryAppend(data, request.AgentCheckPortMapping)
	}
	return
}

func (request *ControlRequest) Read(data []byte) (n int, err error) {
	var id ControlRequestId
	{
		n, err = id.Read(data)
		if data = data[n:]; err != nil {
			return
		}
	}
	switch id {
	case ControlRequestIdPingV2:
		request.Ping = &Ping{}
		n1, err := request.Ping.Read(data)
		if n += n1; err != nil {
			return n, err
		}
	case ControlRequestIdAgentRegisterV1:
		request.AgentRegister = &AgentRegister{}
		n1, err := request.AgentRegister.Read(data)
		if n += n1; err != nil {
			return n, err
		}
	case ControlRequestIdAgentKeepAliveV1:
		request.AgentKeepAlive = &AgentSessionId{}
		n1, err := request.AgentKeepAlive.Read(data)
		if n += n1; err != nil {
			return n, err
		}
	case ControlRequestIdSetupUdpChannelV1:
		request.SetupUdpChannel = &AgentSessionId{}
		n1, err := request.SetupUdpChannel.Read(data)
		if n += n1; err != nil {
			return n, err
		}
	case ControlRequestIdAgentCheckPortMappingV1:
		request.AgentCheckPortMapping = &AgentCheckPortMapping{}
		n1, err := request.AgentCheckPortMapping.Read(data)
		if n += n1; err != nil {
			return n, err
		}
	default:
		err = errors.New("old control request no longer supported")
	}
	return
}
