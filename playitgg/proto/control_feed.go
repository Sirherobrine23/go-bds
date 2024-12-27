package proto

import (
	"encoding/binary"
	"errors"
	"slices"

	"sirherobrine23.com.br/go-bds/go-bds/playitgg/message_encoding"
)

var (
	ErrInvalidFeedID error = errors.New("invalid ControlFeed id")

	_ message_encoding.Binary = &ControlFeed{}
	_ message_encoding.Binary = &NewClient{}
)

type ControlFeed struct {
	NewClient *NewClient
	Response  *ControlRpcMessage[*ControlResponse]
}

func (feed *ControlFeed) Read(data []byte) (int, error) {
	switch binary.BigEndian.Uint32(data[0:4]) {
	case 1:
		feed.Response = &ControlRpcMessage[*ControlResponse]{}
		n, err := feed.Response.Read(data[4:])
		return n + 4, err
	case 2:
		feed.NewClient = &NewClient{}
		n, err := feed.NewClient.Read(data[4:])
		return n + 4, err
	default:
		return 4, ErrInvalidFeedID
	}
}

func (feed ControlFeed) MarshalBinary() ([]byte, error) {
	switch {
	case feed.NewClient != nil:
		feedID := make([]byte, 4)
		binary.BigEndian.PutUint32(feedID, 1)
		data, err := feed.NewClient.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(feedID, data...), nil
	case feed.Response != nil:
		feedID := make([]byte, 4)
		binary.BigEndian.PutUint32(feedID, 2)
		data, err := feed.Response.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(feedID, data...), nil
	default:
		return nil, ErrInvalidFeedID
	}
}

type NewClient struct {
	ConnectAddr      message_encoding.AddrPort
	PeerAddr         message_encoding.AddrPort
	ClaimInstruction ClaimInstructions
	TunnelServerID   uint64
	DataCenterID     uint32
}

func (client NewClient) MarshalBinary() ([]byte, error) {
	ConnectAddrData, err := client.ConnectAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}

	PeerAddrData, err := client.PeerAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}

	claimData, err := client.ClaimInstruction.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return slices.Concat(ConnectAddrData, PeerAddrData, claimData, binary.BigEndian.AppendUint64(nil, client.TunnelServerID), binary.BigEndian.AppendUint32(nil, client.DataCenterID)), nil
}

func (client *NewClient) Read(data []byte) (n int, err error) {
	{
		n1, err := client.ConnectAddr.Read(data)
		n += n1
		data = data[0:n1]
		if err != nil {
			return n, err
		}
	}
	{
		n1, err := client.PeerAddr.Read(data)
		n += n1
		data = data[0:n1]
		if err != nil {
			return n, err
		}
	}
	{
		n1, err := client.ClaimInstruction.Read(data)
		n += n1
		data = data[0:n1]
		if err != nil {
			return n, err
		}
	}
	client.TunnelServerID = binary.BigEndian.Uint64(data[0:4])
	client.DataCenterID = binary.BigEndian.Uint32(data[4:6])
	n += 6
	return
}

type ClaimInstructions struct {
	Address message_encoding.AddrPort
	Token   message_encoding.Buffer
}

func (claim ClaimInstructions) MarshalBinary() ([]byte, error) {
	addrBin, err := claim.Address.MarshalBinary()
	if err != nil {
		return nil, err
	}
	token, _ := claim.Token.MarshalBinary()
	return append(addrBin, token...), nil
}
func (claim *ClaimInstructions) Read(data []byte) (n int, _ error) {
	{
		claim.Address = message_encoding.AddrPort{}
		n1, err := claim.Address.Read(data)
		n += n1
		data = data[:n1]
		if err != nil {
			return n, err
		}
	}
	{
		n1, err := claim.Token.Read(data)
		n += n1
		if err != nil {
			return n, err
		}
	}
	return
}
