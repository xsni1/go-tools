package main

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

type MessageOptions struct {
	id       string
	qdcount  uint16
	ancount  uint16
	nscount  uint16
	arcount  uint16
	question string
	qtype    int
	qclass   int
	// flags
	qr     uint16
	opcode uint16
	aa     uint16
	tc     uint16
	rd     uint16
	ra     uint16
	rcode  uint16
}

type Message struct {
	n int
}

func (m *Message) Encode(opts MessageOptions) []byte {
	// header + qtypes and qclass + length of first and last label + question
	msg := make([]byte, 12+4+1+1+len(opts.question))

	id := uint16(time.Now().Unix())
	binary.BigEndian.PutUint16(msg[0:], id)

	var flags uint16 = 0b0000000000000000
	flags = flags | opts.qr
	flags = flags << 4
    flags = flags | opts.opcode
	flags = flags << 1
	flags = flags | opts.aa
	flags = flags << 1
	flags = flags | opts.tc
	flags = flags << 1
	flags = flags | opts.rd
	flags = flags << 1
	flags = flags | opts.ra
	flags = flags << 3
    // z flag 
    flags = flags << 4
    flags = flags | opts.rcode

    fmt.Printf("flags: %016b\n", flags)
	binary.BigEndian.PutUint16(msg[2:], flags)
	binary.BigEndian.PutUint16(msg[4:], opts.qdcount)
	binary.BigEndian.PutUint16(msg[6:], opts.ancount)
	binary.BigEndian.PutUint16(msg[8:], opts.nscount)
	binary.BigEndian.PutUint16(msg[10:], opts.arcount)

	return msg
}

func main() {
	m := Message{}
    m.Encode(MessageOptions{qr: 1, rd: 1})
	udpaddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
	if err != nil {
		slog.Error("resolving udp addr", "error", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		slog.Error("dialing", "error", err)
		os.Exit(1)
	}

	domain := "dns.google.com"

	// HEADER SECTION
	// header is 12 bytes
	header := make([]byte, 12)
	id := uint16(time.Now().Unix())
	binary.BigEndian.PutUint16(header[0:], id)

	//                          v <= RD flag
	var flags uint16 = 0b0000000100000000
	binary.BigEndian.PutUint16(header[2:], flags)

	// QDCOUNT - number of entries in question section
	var qdcount uint16 = 1
	binary.BigEndian.PutUint16(header[4:], qdcount)

	var ancount uint16 = 0
	binary.BigEndian.PutUint16(header[6:], ancount)

	var nscount uint16 = 0
	binary.BigEndian.PutUint16(header[8:], nscount)

	var arcount uint16 = 0
	binary.BigEndian.PutUint16(header[10:], arcount)

	// QUESTION SECTION
	// domain length + length of first label + closing 0 + two bytes each for qtypes and qclass
	test := make([]byte, len(domain)+4+1+1)
	n := 0
	for _, label := range strings.Split(domain, ".") {
		test[n] = byte(len(label))
		n++
		for i := 0; i < len(label); i++ {
			test[n] = label[i]
			n++
		}
	}
	test[n] = byte(0)
	n++

	var qtype uint16 = 1
	binary.BigEndian.PutUint16(test[n:], qtype)

	var qclass uint16 = 1
	binary.BigEndian.PutUint16(test[n+2:], qclass)

	msg := []byte{}
	msg = append(msg, header...)
	msg = append(msg, test...)

	_, err = conn.WriteTo(msg, udpaddr)
	if err != nil {
		slog.Error("writing to udp socket", "error", err)
	}
	slog.Info("sent", "msg", msg)

	var respbuf [1024]byte
	conn.ReadFrom(respbuf[:])

	if err != nil {
		slog.Error("reading udp msg", "error", err)
	}
	fmt.Println(respbuf)
}
