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

	time.Sleep(16 * time.Millisecond)

	for i := 0; i < 1; i++ {
		pkt := PacketUDP{Sequence: uint16(i + 1), Ack: uint16(i + 1)}

		pn.ReceivePacket(&pkt)
	}

	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("rtt", pn.RTT())

		fmt.Println("pktloss", pn.PacketLoss())
		fmt.Println("ack bandwidth ", pn.AckBandwith())
		fmt.Println("sent bandwidth", pn.SentBandwith())
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
