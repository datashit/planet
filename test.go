package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

func main() {

	ludpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:1455")
	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:1433")

	udp, err := net.DialUDP("udp4", ludpAddr, udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 18)

	sequence := uint32(1)
	session := uint32(0)
	action := 61
	data := byte(22)
	actionType := byte(1)

	binary.LittleEndian.PutUint32(buf[0:], session)
	binary.LittleEndian.PutUint32(buf[4:], sequence)
	binary.LittleEndian.PutUint32(buf[8:], uint32(action))
	buf[12] = actionType
	binary.LittleEndian.PutUint32(buf[13:], 1)
	buf[17] = data

	udp.Write(buf)

	time.Sleep(10 * time.Second)

}
