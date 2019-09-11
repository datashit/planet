package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {

	pn := NewPlayerNetwork()

	go func() {
		for {
			time.Sleep(1 * time.Second)

			pn.update()
		}

	}()

	for i := 0; i < 35; i++ {
		pn.sendPacket()
	}

	for i := 1; i < 20; i++ {
		time.Sleep(16 * time.Millisecond)
		if i == 3 {
			continue
		}

		if i == 10 {
			continue
		}
		now := time.Now()
		pkt := PacketUDP{Sequence: uint16(i), Ack: uint16(i)}

		pn.ReceivePacket(&pkt, now)
	}

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("rtt", pn.RTT())

		fmt.Println("pktloss", pn.PacketLoss())
		fmt.Println("ack bandwidth ", pn.AckBandwidth())
		fmt.Println("sent bandwidth", pn.SentBandwidth())
		fmt.Println("recv bandwidth", pn.RecvBandwidth())
	}

	return

	lp, err := net.ListenPacket("udp", ":1433")
	if err != nil {
		log.Fatalln(err)
	}

	var buf [8]byte
	for {
		n, addr, err := lp.ReadFrom(buf[:])
		if err != nil {
			continue
		}
		if n < 8 {
			continue
		}

		parsePacket(buf[:n])
		handlePacket(addr)

	}
}

func parsePacket(buf []byte) {

}

func handlePacket(addr net.Addr) {

}

func sendPacketUDP(conn net.PacketConn, packet *PacketUDP) {

}
