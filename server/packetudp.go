package main

type PacketUDP struct {
	Protocol    uint16
	Sequence    uint16
	Ack         uint16
	AckBitfield uint32
	DataSize    uint16
	Data        []byte
}
