package main

import (
	"fmt"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Workiva/go-datastructures/bitarray"
	_ "github.com/Workiva/go-datastructures/bitarray"
	"github.com/cornelk/hashmap"
)

const (
	AckSize = 32
)

type PlayerNetwork struct {
	sync.RWMutex
	localSequence     uint32
	remoteSequence    atomic.Value
	remoteAckBitfield bitarray.Bitmap32
	conn              net.PacketConn
	rtt               atomic.Value
	rttmap            hashmap.HashMap
	ackqueue          hashmap.HashMap
}

func NewPlayerNetwork() *PlayerNetwork {

	pn := PlayerNetwork{}

	pn.remoteSequence.Store(uint16(0))
	pn.rtt.Store(float32(0))

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

func (p *PlayerNetwork) incrSeq() uint16 {

	return uint16(atomic.AddUint32(&p.localSequence, 1))
}

func (p *PlayerNetwork) getLocalSequence() uint16 {

	return uint16(atomic.LoadUint32(&p.localSequence))

}

func (p *PlayerNetwork) ack(last uint16, bitmap uint32) {
	now := time.Now()

	p.ackqueue.Insert(last, now)

	bmap := bitarray.Bitmap32(bitmap)

	for i := 0; i < AckSize; i++ {
		bitField := i + 1
		seq := last - uint16(bitField)

		if bmap.GetBit(uint(bitField)) {
			fmt.Println(6)
			p.ackqueue.Insert(seq, now)
		}

	}
}

func (p *PlayerNetwork) sendPacket() {
	sequence := p.incrSeq()

	pkt := PacketUDP{Sequence: sequence}
	pkt.Ack = p.RemoteSequence()

	pkt.AckBitfield = uint32(p.remoteAckBitfield)
	// Sendpacket

	// RTT calc process
	p.rttmap.Set(sequence, time.Now())
}

func (p *PlayerNetwork) impRTT(newrtt float32) {

	rtt := p.RTT()
	rtt = rtt + (0.10 * (newrtt - rtt))
	p.rtt.Store(rtt)
}

func (p *PlayerNetwork) ReceivePacket(pkt *PacketUDP) {

	invalidPacket := false
	// Remote Ack
	remoteSeq := p.RemoteSequence()
	if validSequence(remoteSeq, pkt.Sequence) {
		shift := uint(math.Abs(float64(pkt.Sequence - remoteSeq)))

		p.setRemoteSequence(pkt.Sequence)
		setBitfield := uint(pkt.Sequence - remoteSeq)
		p.remoteAckBitfield = bitarray.Bitmap32((p.remoteAckBitfield << shift)).SetBit(setBitfield)
		fmt.Printf("% 032b", p.remoteAckBitfield)
		fmt.Println()
	} else if (remoteSeq-AckSize) >= pkt.Sequence && pkt.Sequence <= remoteSeq {

		setBitfield := uint(remoteSeq - pkt.Sequence)
		if p.remoteAckBitfield.GetBit(setBitfield) {
			return // Duplicate Packet
		}
		p.remoteAckBitfield = p.remoteAckBitfield.SetBit(setBitfield)
		fmt.Println("remoteAckBitfield", p.remoteAckBitfield)

	} else {
		invalidPacket = true
	}
	// Remote Ack END

	//Local Ack
	p.ack(pkt.Ack, pkt.AckBitfield)
	//Local Ack END

	if invalidPacket {
		return
	}
}

func (p *PlayerNetwork) ackRTT() {

	now := time.Now()

	for val := range p.rttmap.Iter() {

		sendTime := val.Value.(time.Time)
		if income, ok := p.ackqueue.Get(val.Key); ok {
			come := income.(time.Time)

			p.impRTT(float32(come.Sub(sendTime) / time.Millisecond))
			p.rttmap.Del(val.Key)
			p.ackqueue.Del(val.Key)
			continue
		}

		if now.Sub(sendTime) > 1*time.Second {
			// Packet Loss
			p.rttmap.Del(val.Key)
		}
	}

}
