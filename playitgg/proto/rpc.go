package proto

import (
	"encoding/binary"
	"io"

	"sirherobrine23.com.br/go-bds/go-bds/playitgg/message_encoding"
)

var _ message_encoding.Binary = &ControlRpcMessage[message_encoding.Binary]{}

type ControlRpcMessage[T message_encoding.Binary] struct {
	RequestID uint64
	Content   T
}

func (rpc ControlRpcMessage[T]) MarshalBinary() ([]byte, error) {
	contentBuff, err := rpc.Content.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, rpc.RequestID)
	return append(buff, contentBuff...), nil
}

func (rpc *ControlRpcMessage[T]) Read(data []byte) (int, error) {
	if len(data) < 8 {
		return 0, io.ErrUnexpectedEOF
	}
	rpc.RequestID = binary.BigEndian.Uint64(data[0:8])
	n, err := rpc.Content.Read(data[8:])
	return n + 8, err
}
