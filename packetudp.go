package planet

const DATA_SIZE_BYTES = 15
const PACKET_UDP_HEADER_BYTES = 17

type PacketUDP struct {
	Session     uint32
	Protocol    uint16
	Sequence    uint16
	Ack         uint16
	AckBitfield uint32
	DataType    uint8
	DataSize    uint16
	Data        []byte
}

func (p *PacketUDP) Write() []byte {

	buffer := NewBuffer(PACKET_UDP_HEADER_BYTES + int(p.DataSize))

	buffer.WriteUint32(p.Session)
	buffer.WriteUint16(p.Protocol)
	buffer.WriteUint16(p.Sequence)
	buffer.WriteUint16(p.Ack)
	buffer.WriteUint32(p.AckBitfield)
	buffer.WriteUint8(p.DataType)
	buffer.WriteUint16(p.DataSize)
	buffer.WriteBytes(p.Data)

	return buffer.Bytes()

}

func NewPacket(buf []byte) *PacketUDP {
	if len(buf) < PACKET_UDP_HEADER_BYTES {
		return nil
	}

	packetBuffer := NewBufferFromRef(buf)

	_, err := packetBuffer.GetBytes(DATA_SIZE_BYTES)
	if err != nil {
		return nil
	}

	dataSize, err := packetBuffer.GetUint16()
	if err != nil {
		return nil
	}

	data, err := packetBuffer.GetBytes(int(dataSize))
	if err != nil {
		return nil
	}

	packetBuffer.Reset()

	pckt := PacketUDP{DataSize: dataSize, Data: data}

	pckt.Session, err = packetBuffer.GetUint32()
	if err != nil {
		return nil
	}

	pckt.Protocol, err = packetBuffer.GetUint16()
	if err != nil {
		return nil
	}

	pckt.Sequence, err = packetBuffer.GetUint16()
	if err != nil {
		return nil
	}

	pckt.Ack, err = packetBuffer.GetUint16()
	if err != nil {
		return nil
	}
	pckt.AckBitfield, err = packetBuffer.GetUint32()
	if err != nil {
		return nil
	}
	pckt.DataType, err = packetBuffer.GetUint8()
	if err != nil {
		return nil
	}

	return &pckt

}
