package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
)

// Peer represents a Peer
type Peer struct {
	IP             net.IP
	Port           int
	Socket         net.Conn
	BytesLeft      *bytes.Buffer
	AMPeerID       [20]byte
	InfoHash       [20]byte
	AMChoking      bool
	AMInterested   bool
	PeerChoking    bool
	PeerInterested bool
}

// PeerHandshake defines the message to send to a Peer to initiate a connection
type PeerHandshake struct {
	Pstrlen  byte
	Pstr     [19]byte
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

// Connect initializes the connection to the peer
func (p *Peer) Connect() (err error) {
	addr := p.IP.String() + ":" + strconv.Itoa(p.Port)
	p.Socket, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	//fmt.Println("new peer:", addr)
	err = p.handshake()
	if err != nil {
		return err
	}
	fmt.Println("new peer:", addr)
	return nil
}

func (p *Peer) write(b interface{}) (err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, b)
	if err != nil {
		return err
	}
	n, err := p.Socket.Write(buf.Bytes())
	if n != buf.Len() {
		return errors.New("did not send everything to socket")
	}
	return nil
}

func (p *Peer) read(bb *bytes.Buffer) (err error) {
	b := make([]byte, 4096)
	_, err = p.Socket.Read(b)
	if err != nil {
		return err
	}
	bb.Write(b)
	return nil
}

// LoopReceive handles new messages
func (p *Peer) LoopReceive() (err error) {
	// every message starts with 4 bytes
	if p.BytesLeft.Len() < 4 {
		err = p.read(p.BytesLeft)
		return
	}
	rawLen := p.BytesLeft.Next(4) // read 4 bytes => len field
	len := binary.BigEndian.Uint32(rawLen)
	switch len {
	case 0:
		//keep alive
		//fmt.Println("KEEP_ALIVE")
		return nil
	case 1:
		rawID, err := p.BytesLeft.ReadByte()
		if err != nil {
			return err
		}
		switch rawID {
		case 0:
			fmt.Println("CHOKE")
			return nil
		case 1:
			fmt.Println("UNCHOKE")
			return nil
		case 2:
			fmt.Println("INTERESTED")
			return nil
		case 3:
			fmt.Println("NOT_INTERESTED")
			return nil
		default:
			fmt.Printf("received message:  %08b %08b %s\n", len, rawID, p.IP.String())
			return nil
		}
	case 5:
		//fmt.Println("HAVE")
		rawID, err := p.BytesLeft.ReadByte()
		if err != nil {
			return err
		}
		if rawID != 4 {
			fmt.Printf("HAVE NOT RECOGNIZED: %08b", rawID)
		}
		return nil
	default:
		fmt.Printf("received message: [% x] %08b %s\n", rawLen, rawLen, p.IP.String())
	}

	return nil
}

func (p *Peer) handshake() (err error) {
	phr := PeerHandshake{
		Pstrlen: byte(19),
		//Pstr:     [19]byte(string("")),
		Reserved: [8]byte{0, 0, 0, 0, 0, 16, 0, 0},
		// we support extensions and nothing else
		// https://wiki.theory.org/index.php/BitTorrentSpecification
		InfoHash: p.InfoHash,
		PeerID:   p.AMPeerID,
	}
	copy(phr.Pstr[:], []byte("BitTorrent protocol"))
	err = p.write(&phr)
	if err != nil {
		return err
	}

	/* read section */
	err = p.read(p.BytesLeft)
	if err != nil {
		return err
	}
	var respPhr PeerHandshake
	err = binary.Read(p.BytesLeft, binary.BigEndian, &respPhr)
	if err != nil {
		return err
	}
	return nil
}
