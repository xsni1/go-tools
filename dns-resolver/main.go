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


func main() {
    udpaddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
    if err != nil {
		slog.Error("resolving udp addr", "error", err)
		os.Exit(1)
    }
	conn, err := net.DialUDP("udp", nil, udpaddr)
	// conn, err := net.Dial("udp", "8.8.8.8:53")
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
    var name string
    for _, label := range strings.Split(domain, ".") {
        name += fmt.Sprint(len(label))
        name += label
    }
    name += "0"

    question := make([]byte, len(name) + 4)
    // Strings in go are by default encoded in utf-8, for ASCII characters they are 1 byte long
    for i := 0; i < len(name); i++ {
        question[i] = name[i]
    }

    var qtype uint16 = 0
    binary.BigEndian.PutUint16(question[len(name):], qtype)

    var qclass uint16 = 0
    binary.BigEndian.PutUint16(question[len(name)+2:], qclass)

    fmt.Println(header)
    fmt.Println(name)
    fmt.Println(question)

    msg := []byte{}
    msg = append(msg, header...)
    msg = append(msg, question...)

    n, err := conn.Write(msg)
    if err != nil {
        slog.Error("writing to udp socket", "error", err)
    }
    slog.Info("sent", "msg", msg, "n", n)

    var respbuf [1024]byte
    conn.ReadFrom(respbuf[:])
    // resp, err := bufio.NewReader(conn).ReadBytes
    if err != nil {
        slog.Error("reading udp msg", "error", err)
    }
    fmt.Println(respbuf)
}
