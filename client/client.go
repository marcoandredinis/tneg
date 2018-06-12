package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/marcoandredinis/tneg/metadata"
)

// Client represents a Client instance
type Client struct {
	torrent      metadata.Torrent
	fileLocation string
	listenPort   uint16
	peerID       [20]byte
	trackers     []TrackerConnection
	peers        map[string]Peer
}

// FromTorrent creates a Client from a metadata.Torrent and a Path (file destination)
func FromTorrent(t metadata.Torrent) (c Client, err error) {
	if len(t.Trackers) < 1 {
		return Client{}, fmt.Errorf("we need at least one tracker")
	}
	c.torrent = t
	return c, nil
}

// StartTorrent is
func (c *Client) StartTorrent() (err error) {
	copy(c.peerID[:], []byte("-TG0001.random_value")) // TODO refactor to get a random value?
	c.listenPort = uint16(28288)                      // TODO refactor to get a valid port? UPnP
	c.fileLocation = "~/Downloads/"                   // TODO refactor to set custom folder
	c.peers = make(map[string]Peer)

	// create tracker connection
	for _, t := range c.torrent.Trackers {
		tc := TrackerConnection{
			URL:                  *t,
			LastWaitedTime:       0,
			SleepBetweenAnnounce: 15,
			InfoHash:             c.torrent.HashValue,
			ListenPort:           c.listenPort,
			PeerID:               c.peerID,
		}
		c.trackers = append(c.trackers, tc)
	}

	chPeers := make(chan PeerResponse)

	for _, ctc := range c.trackers {
		go func(ctc TrackerConnection) {
			err = ctc.Connect()
			if err != nil {
				return
			}
			err = ctc.Announce()
			if err != nil {
				return
			}
			for _, p := range ctc.Peers {
				chPeers <- p
			}
		}(ctc)
	}
	// Connect to each Peer and ask for pieces

	for {
		pr := <-chPeers
		p := Peer{
			IP:             make(net.IP, 4),
			Port:           int(pr.Port),
			AMPeerID:       c.peerID,
			InfoHash:       c.torrent.HashValue,
			BytesLeft:      new(bytes.Buffer),
			AMChoking:      true,
			AMInterested:   false,
			PeerChoking:    true,
			PeerInterested: false,
		}
		binary.BigEndian.PutUint32(p.IP, pr.IP)
		ipPort := p.IP.String() + ":" + strconv.Itoa(p.Port)
		if _, ok := c.peers[ipPort]; ok {
			//fmt.Println("peer repeated", ipPort)
			continue
		}
		c.peers[ipPort] = p
		go func() {
			err = p.Connect()
			if err != nil {
				//fmt.Println("erro [", p.IP, ":", p.Port, "]:", err)
				return
			}
			go func() {
				for {
					err = p.LoopReceive()
					if err != nil {
						//fmt.Println("erro [", p.IP, ":", p.Port, "]:", err)
						return
					}
				}
			}()
		}()
		//time.Sleep(time.Second * 15)
	}
	//return nil

}
