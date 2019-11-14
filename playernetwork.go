package planet

import (
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/Workiva/go-datastructures/bitarray"
	_ "github.com/Workiva/go-datastructures/bitarray"
	"github.com/cornelk/hashmap"
)

const (
	AckSize                   = 32
	RttSmoothingFactor        = 0.0025
	PacketLossSmoothingFactor = 0.20
	BandwidthSmoothingFactor  = 0.10
)

type PlayerNetwork struct {
	localSequence     uint32
	remoteSequence    atomic.Value
	remoteAckBitfield bitarray.Bitmap32
	rtt               atomic.Value
	packetloss        atomic.Value
	sentpackets       hashmap.HashMap
	recvpackets       hashmap.HashMap
	ackbandwidth      atomic.Value
	sentbandwidth     atomic.Value
	recvbandwidth     atomic.Value
	conn              *net.UDPConn
}

type PacketAck struct {
	Bytes    uint
	SentTime time.Time
	RecvTime time.Time
	Acked    bool
}

func NewPlayerNetwork(conn *net.UDPConn) *PlayerNetwork {

	pn := PlayerNetwork{conn: conn}

	pn.remoteSequence.Store(uint16(0))
	pn.rtt.Store(float32(0))
	pn.packetloss.Store(float32(0))
	pn.ackbandwidth.Store(float32(0))
	pn.sentbandwidth.Store(float32(0))
	pn.recvbandwidth.Store(float32(0))

	go pn.update()

	return &pn
}

func (p *PlayerNetwork) setRemoteSequence(seq uint16) {
	p.remoteSequence.Store(seq)

}
func (p *PlayerNetwork) RemoteSequence() uint16 {
	return p.remoteSequence.Load().(uint16)
}
func (p *PlayerNetwork) RTT() float32 {
	return p.rtt.Load().(float32)
}
func (p *PlayerNetwork) SentBandwidth() float32 {
	return p.sentbandwidth.Load().(float32)
}

func (p *PlayerNetwork) RecvBandwidth() float32 {
	return p.recvbandwidth.Load().(float32)
}

func (p *PlayerNetwork) AckBandwidth() float32 {
	return p.ackbandwidth.Load().(float32)
}

func (p *PlayerNetwork) PacketLoss() float32 {
	return p.packetloss.Load().(float32)
}

func (p *PlayerNetwork) incrSeq() uint16 {

	return uint16(atomic.AddUint32(&p.localSequence, 1))
}

func (p *PlayerNetwork) getLocalSequence() uint16 {

	return uint16(atomic.LoadUint32(&p.localSequence))

}

func (p *PlayerNetwork) ack(last uint16, bitmap uint32, now time.Time) {

	bifField := bitarray.Bitmap32(bitmap)
	for i := 0; i < AckSize; i++ {

		if !bifField.GetBit(uint(i)) {
			continue
		}

		ackSeq := last - uint16(i)

		if val, ok := p.sentpackets.Get(ackSeq); ok {

			pktack := val.(PacketAck)
			if !pktack.Acked {

				pktack.Acked = true
				pktack.RecvTime = now

				p.sentpackets.Set(ackSeq, pktack)

				p.impRTT(float32(pktack.RecvTime.Sub(pktack.SentTime) / time.Millisecond))
			}
		}
	}
}

func (p *PlayerNetwork) generateSendPacket(protocol uint16, datatype uint8, data []byte) *PacketUDP {
	sequence := p.incrSeq()

	pkt := PacketUDP{Sequence: sequence, Protocol: protocol, DataType: datatype, Data: data, DataSize: uint16(len(data))}
	pkt.Ack = p.RemoteSequence()

	pkt.AckBitfield = uint32(p.remoteAckBitfield)

	return &pkt
}

func (p *PlayerNetwork) SendPacket(protocol uint16, datatype uint8, data []byte) error {

	pkt := p.generateSendPacket(protocol, datatype, data)
	// Sendpacket

	_, err := p.conn.Write(pkt.Write())
	if err == nil {
		p.sentpackets.Set(pkt.Sequence, PacketAck{SentTime: time.Now(), Bytes: uint(12 + pkt.DataSize)})
	}
	return err

}

func (p *PlayerNetwork) SendPacketToUDP(protocol uint16, datatype uint8, data []byte, addr *net.UDPAddr) error {
	pkt := p.generateSendPacket(protocol, datatype, data)
	// Sendpacket

	_, err := p.conn.WriteToUDP(pkt.Write(), addr)
	if err == nil {
		p.sentpackets.Set(pkt.Sequence, PacketAck{SentTime: time.Now(), Bytes: uint(12 + pkt.DataSize)})
	}
	return err

}

func (p *PlayerNetwork) impRTT(newrtt float32) {
	rtt := p.RTT()
	if (rtt == 0.0 && newrtt > 0.0) || math.Abs(float64(rtt-newrtt)) < 0.00001 {
		rtt = newrtt
	} else {
		rtt += (newrtt - rtt) * RttSmoothingFactor
	}
	p.rtt.Store(rtt)
}

func (p *PlayerNetwork) updatePacketLoss(newpktloss float32) {

	pktloss := p.PacketLoss()

	if math.Abs(float64(pktloss-newpktloss)) > 0.1 {
		pktloss += (newpktloss - pktloss) * PacketLossSmoothingFactor
	} else {
		pktloss = newpktloss
	}

	p.packetloss.Store(pktloss)
}

func (p *PlayerNetwork) updateRecvBandWidth(newbandwidth float32) {
	bandwidth := p.RecvBandwidth()

	if math.Abs(float64(bandwidth-newbandwidth)) > 0.1 {
		bandwidth += (newbandwidth - bandwidth) * BandwidthSmoothingFactor
	} else {
		bandwidth = newbandwidth
	}

	p.recvbandwidth.Store(bandwidth)

}

func (p *PlayerNetwork) updateSentBandWidth(newbandwidth float32) {
	bandwidth := p.SentBandwidth()

	if math.Abs(float64(bandwidth-newbandwidth)) > 0.1 {
		bandwidth += (newbandwidth - bandwidth) * BandwidthSmoothingFactor
	} else {
		bandwidth = newbandwidth
	}

	p.sentbandwidth.Store(bandwidth)

}

func (p *PlayerNetwork) updateAckBandWidth(newbandwidth float32) {
	bandwidth := p.AckBandwidth()

	if math.Abs(float64(bandwidth-newbandwidth)) > 0.00001 {
		bandwidth += (newbandwidth - bandwidth) * BandwidthSmoothingFactor
	} else {
		bandwidth = newbandwidth
	}

	p.ackbandwidth.Store(bandwidth)

}

func (p *PlayerNetwork) ReceivePacket(pkt *PacketUDP, recvTime time.Time, addr *net.UDPAddr, process func(pkt *PacketUDP, addr *net.UDPAddr)) {

	// Remote Ack
	remoteSeq := p.RemoteSequence()
	if !validSequence(remoteSeq, pkt.Sequence) {
		return
	}

	shift := uint(math.Abs(float64(pkt.Sequence - remoteSeq)))
	p.remoteAckBitfield = bitarray.Bitmap32((p.remoteAckBitfield << shift))
	p.setRemoteSequence(pkt.Sequence)

	setBitfield := uint((pkt.Sequence - 1) - remoteSeq)

	p.remoteAckBitfield = p.remoteAckBitfield.SetBit(setBitfield)
	// Remote Ack END

	// Packet Process
	process(pkt, addr)
	// Packet Process END

	//Local Ack
	p.ack(pkt.Ack, pkt.AckBitfield, recvTime)
	//Local Ack END

	p.recvpackets.Set(pkt.Sequence, PacketAck{RecvTime: recvTime, Bytes: uint(12 + pkt.DataSize)})
	// if invalidPacket {
	// 	return
	// }

}

func (p *PlayerNetwork) calculateSentAndPacketLoss() {
	sendLen := p.sentpackets.Len()
	if sendLen == 0 {
		p.updatePacketLoss(0)
		p.updateAckBandWidth(0)
		p.updateSentBandWidth(0)
		return
	}

	maxTime := time.Unix(1<<63-62135596801, 999999999)

	loss := 0
	bytesSentACK := uint(0)
	startTimeACK := maxTime
	finishTimeACK := time.Time{}

	bytesSentSent := uint(0)
	startTimeSent := maxTime
	finishTimeSent := time.Time{}

	sendLen = 0

	now := time.Now()
	for val := range p.sentpackets.Iter() {

		pktack := val.Value.(PacketAck)

		if now.Sub(pktack.SentTime) < 1*time.Second {
			continue
		}

		sendLen++

		if !pktack.Acked {
			loss++
			p.sentpackets.Del(val.Key)
		}

		if pktack.Acked {
			p.sentpackets.Del(val.Key)
			bytesSentACK += pktack.Bytes
			if pktack.SentTime.Before(startTimeACK) {
				startTimeACK = pktack.SentTime
			}

			if pktack.SentTime.After(finishTimeACK) {
				finishTimeACK = pktack.SentTime
			}
		}

		bytesSentSent += pktack.Bytes
		if pktack.SentTime.Before(startTimeSent) {
			startTimeSent = pktack.SentTime
		}

		if pktack.SentTime.After(finishTimeSent) {
			finishTimeSent = pktack.SentTime
		}

	}

	pktLoss := (float32(loss) / float32(sendLen)) * 100.0
	p.updatePacketLoss(pktLoss)

	if !startTimeACK.Equal(maxTime) && !finishTimeACK.IsZero() {

		t := finishTimeACK.Sub(startTimeACK).Seconds()
		newbandwidth := float32(float64(bytesSentACK*8) / (float64(t) * 1000))

		p.updateAckBandWidth(newbandwidth)
	}

	if !startTimeSent.Equal(maxTime) && !finishTimeSent.IsZero() {
		t := finishTimeSent.Sub(startTimeSent).Seconds()
		newbandwidth := float32(float64(bytesSentSent*8) / (float64(t) * 1000))

		p.updateSentBandWidth(newbandwidth)
	}
}

func (p *PlayerNetwork) calculateRecvBandWidth() {

	recvLen := p.recvpackets.Len()
	if recvLen == 0 {
		p.updateRecvBandWidth(0)
		return
	}
	maxTime := time.Unix(1<<63-62135596801, 999999999)
	bytes := uint(0)
	startTime := maxTime
	finishTime := time.Time{}

	for val := range p.recvpackets.Iter() {

		pktack := val.Value.(PacketAck)

		p.recvpackets.Del(val.Key)
		bytes += pktack.Bytes
		if pktack.RecvTime.Before(startTime) {
			startTime = pktack.RecvTime
		}

		if pktack.RecvTime.After(finishTime) {
			finishTime = pktack.RecvTime
		}

	}

	if !startTime.Equal(maxTime) && !finishTime.IsZero() {
		t := finishTime.Sub(startTime).Seconds()

		newbandwidth := float32((float64(bytes) * 8) / (float64(t) * 1000))
		p.updateRecvBandWidth(newbandwidth)
	}

}

func (p *PlayerNetwork) update() {
	for {
		time.Sleep(1 * time.Second)

		p.calculateSentAndPacketLoss()
		p.calculateRecvBandWidth()
	}

}
