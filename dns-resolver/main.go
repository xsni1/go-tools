package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Question struct {
	name   string
	qtype  uint16
	qclass uint16
}

type Nameserver struct {
	name   string
	nstype uint16
	class  uint16
	ttl    uint32
	rdlen  uint16
	rdata  string
}

type AdditionalResource struct {
	name   string
	nstype uint16
	class  uint16
	ttl    uint32
	rdlen  uint16
	rdata  string
}

type AnswerResource struct {
	name   string
	nstype uint16
	class  uint16
	ttl    uint32
	rdlen  uint16
	rdata  string
}

type Message struct {
	id       uint16
	qdcount  uint16
	ancount  uint16
	nscount  uint16
	arcount  uint16
	question string
	qtype    uint16
	qclass   uint16
	// flags
	qr     uint16
	opcode uint16
	aa     uint16
	tc     uint16
	rd     uint16
	ra     uint16
	rcode  uint16

	// response only for now
	questions   []Question
	nameservers []Nameserver
	additional  []AdditionalResource
	answers     []AnswerResource
}

func encodeDnsMsg(opts Message) []byte {
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
	flags = flags << 4
	flags = flags | opts.rcode

	binary.BigEndian.PutUint16(msg[2:], flags)
	binary.BigEndian.PutUint16(msg[4:], opts.qdcount)
	binary.BigEndian.PutUint16(msg[6:], opts.ancount)
	binary.BigEndian.PutUint16(msg[8:], opts.nscount)
	binary.BigEndian.PutUint16(msg[10:], opts.arcount)

	// question section
	n := 12
	for _, label := range strings.Split(opts.question, ".") {
		msg[n] = byte(len(label))
		n++
		for i := 0; i < len(label); i++ {
			msg[n] = label[i]
			n++
		}
	}
	msg[n] = byte(0)
	n++

	binary.BigEndian.PutUint16(msg[n:], opts.qtype)
	binary.BigEndian.PutUint16(msg[n+2:], opts.qclass)

	return msg
}

func decodeDnsMsg(buffer []byte) Message {
	msg := Message{}

	msg.id = binary.BigEndian.Uint16(buffer[0:2])
	msg.qdcount = binary.BigEndian.Uint16(buffer[4:6])
	msg.nscount = binary.BigEndian.Uint16(buffer[8:10])
	msg.ancount = binary.BigEndian.Uint16(buffer[6:8])
	msg.arcount = binary.BigEndian.Uint16(buffer[10:12])

	tmp := ""
	n := 12

	for i := 0; i < int(msg.qdcount); i++ {
		for {
			labellen := int(buffer[n])
			if labellen == 0 {
				n++
				break
			}
			n++
			tmp += string(buffer[n : n+labellen])
			n += labellen
		}

		q := Question{
			name:   tmp,
			qtype:  binary.BigEndian.Uint16(buffer[n : n+2]),
			qclass: binary.BigEndian.Uint16(buffer[n+2 : n+4]),
		}
		n += 4
		msg.questions = append(msg.questions, q)
	}

	for i := 0; i < int(msg.ancount); i++ {
		name, newn := parseName(buffer, n)
		n = newn
		nstype := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		class := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		ttl := binary.BigEndian.Uint32(buffer[n : n+4])
		n += 4
		datalen := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		rdata := buffer[n : n+int(datalen)]
		n += int(datalen)

		address := parseAddress(uint8(nstype), rdata)

		ns := AnswerResource{
			name:   name,
			nstype: nstype,
			class:  class,
			ttl:    ttl,
			rdlen:  datalen,
			rdata:  address,
		}

		msg.answers = append(msg.answers, ns)
	}

	for i := 0; i < int(msg.nscount); i++ {
		name, newn := parseName(buffer, n)
		n = newn
		nstype := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		class := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		ttl := binary.BigEndian.Uint32(buffer[n : n+4])
		n += 4
		datalen := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		rdata, newn := parseName(buffer, n)
		n = newn

		fmt.Println(name, nstype, class, ttl, datalen, rdata)
		ns := Nameserver{
			name:   name,
			nstype: nstype,
			class:  class,
			ttl:    ttl,
			rdlen:  datalen,
			rdata:  rdata,
		}

		msg.nameservers = append(msg.nameservers, ns)
	}

	for i := 0; i < int(msg.arcount); i++ {
		name, newn := parseName(buffer, n)
		n = newn
		nstype := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		class := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		ttl := binary.BigEndian.Uint32(buffer[n : n+4])
		n += 4
		datalen := binary.BigEndian.Uint16(buffer[n : n+2])
		n += 2
		rdata := buffer[n : n+int(datalen)]
		n += int(datalen)

		address := parseAddress(uint8(nstype), rdata)

		addres := AdditionalResource{
			name:   name,
			nstype: nstype,
			class:  class,
			ttl:    ttl,
			rdlen:  datalen,
			rdata:  address,
		}

		msg.additional = append(msg.additional, addres)
	}

	return msg
}

func main() {
    query := flag.String("q", "", "address to lookup")
    flag.Parse()
	roots := []string{
		"a.root-servers.net",
		"b.root-servers.net",
		"c.root-servers.net",
		"d.root-servers.net",
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Printf("listen udp: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	recursiveQuery(roots, conn, *query)

}

func recursiveQuery(nameservers []string, conn *net.UDPConn, query string) bool {
	fmt.Println("rec query", nameservers)
	for _, ns := range nameservers {
		encodedmsg := encodeDnsMsg(Message{
			rd:       1,
			qtype:    1,
			qclass:   1,
			qdcount:  1,
			question: query,
		})

		ns = fmt.Sprintf("%s:53", ns)
		fmt.Println("CALl: ", ns)
		udpaddr, err := net.ResolveUDPAddr("udp", ns)
		_, err = conn.WriteTo(encodedmsg, udpaddr)
		if err != nil {
			slog.Error("writing to udp socket", "error", err)
			break
		}
		fmt.Println("message sent: ", encodedmsg)

		var respbuf [1024]byte
		_, _, err = conn.ReadFrom(respbuf[:])
		if err != nil {
			slog.Error("reading udp msg", "error", err)
			break
		}

		decodedmsg := decodeDnsMsg(respbuf[:])
		fmt.Printf("decoded msg: %+v\n", decodedmsg)

		if decodedmsg.ancount > 0 {
			for _, v := range decodedmsg.answers {
				fmt.Printf("found: %s\n", v.rdata)
			}
			return true
		}

		if decodedmsg.nscount > 0 {
			fmt.Println("szukamy dalej")
			addrs := []string{}

			for _, v := range decodedmsg.nameservers {
				// nameserver
				if v.nstype == 2 {
					addrs = append(addrs, v.rdata)
				}
			}

			return recursiveQuery(addrs, conn, query)
		}
	}

	return false
}

func parseName(buffer []byte, n int) (string, int) {
	name := ""
	// If compression reserved bits present
	for {
		if buffer[n]&0b11000000 > 0 {
			offset := binary.BigEndian.Uint16(buffer[n : n+2])
			offset = offset - 1<<15 - 1<<14
			n += 2

			p, _ := parseName(buffer, int(offset))
			name += p
			return name, n
		} else {
			labellen := int(buffer[n])
			n++
			if labellen == 0 {
				if name[len(name)-1] == '.' {
					name = name[:len(name)-1]
				}
				return name, n
			}
			name += fmt.Sprintf("%s.", string(buffer[n:n+labellen]))
			n += labellen
		}
	}
}

func parseAddress(nstype uint8, data []byte) string {
	address := ""
	// ipv4
	if nstype == 1 {
		for idx, v := range data {
			if idx == len(data)-1 {
				address += strconv.Itoa(int(v))
				continue
			}

			address += fmt.Sprintf("%s.", strconv.Itoa(int(v)))
		}
		// ipv6
	} else if nstype == 28 {
		for i := 0; i < len(data); i += 2 {
			if i == len(data)-1 {
				address += fmt.Sprintf("%02x%02x", data[i], data[i+1])
			}

			address += fmt.Sprintf("%02x%02x:", data[i], data[i+1])
		}
	}

	return address
}
