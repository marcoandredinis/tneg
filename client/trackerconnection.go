package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

// TrackerConnection represents a connection to a Tracker
type TrackerConnection struct {
	URL                  url.URL
	Socket               net.Conn
	ConnectionID         int64
	TransactionID        int32
	SleepBetweenAnnounce int32
	LastWaitedTime       int
	PeerID               [20]byte
	InfoHash             [20]byte
	Downloaded           int64
	Left                 int64
	Uploaded             int64
	ListenPort           uint16
	Peers                []PeerResponse
}

// PeerResponse represents a Peer
type PeerResponse struct {
	IP   uint32
	Port uint16
}

// Event is
type Event int32

// Event is
const (
	EventNone Event = iota
	EventCompleted
	EventStarted
	EventStopped
)

// Action is
type Action int32

// Action is
const (
	ActionConnect Action = iota
	ActionAnnounce
	ActionScrap
	ActionError
)

const (
	connectionIDInitial = 0x41727101980
)

// HeaderRequest contains common data for every Request call
type HeaderRequest struct {
	ConnectionID  int64
	Action        Action
	TransactionID int32
}

// HeaderResponse contains common data for every Response
type HeaderResponse struct {
	Action        Action
	TransactionID int32
}

// ConnectingBodyResponse contains the connectionID
type ConnectingBodyResponse struct {
	ConnectionID int64
}

// AnnouncingBodyRequest contains the body fields received after an AnnouncingRequest
type AnnouncingBodyRequest struct {
	InfoHash   [20]byte
	PeerID     [20]byte
	Downloaded int64
	Left       int64
	Uploaded   int64
	Event      Event
	IP         uint32
	Key        uint32
	NumWant    int32
	Port       uint16
	Extensions uint16
}

// AnnouncingBodyResponse is
type AnnouncingBodyResponse struct {
	Interval int32
	Leechers int32
	Seeders  int32
}

func (tc *TrackerConnection) write(h interface{}, b interface{}) (err error) {
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, h)
	if err != nil {
		return err
	}
	if b != nil {
		err = binary.Write(&buf, binary.BigEndian, b)
		if err != nil {
			return err
		}
	}
	n, err := tc.Socket.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buf.Len() {
		return errors.New("did not send everything to socket")
	}
	return nil
}

func (tc *TrackerConnection) read(hresp *HeaderResponse, bresp interface{}) (bytesLeft *bytes.Buffer, err error) {
	b := make([]byte, 40960)

	// set read deadline to 15 seconds (and try at most 4 times)
	err = tc.Socket.SetReadDeadline(time.Now().Add(time.Second * 15))
	if err != nil {
		return nil, err
	}
	tc.LastWaitedTime = 15

	for {
		n, err := tc.Socket.Read(b)
		if opE, ok := err.(*net.OpError); ok {
			if opE.Timeout() {
				if tc.LastWaitedTime == 15*4 {
					return nil, errors.New("[TrackerConnection] error: giving up on tracker " + tc.URL.Host)
				}
				tc.LastWaitedTime += 15
				time.Sleep(time.Second * 15)
				continue
			}
			break
		}
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(b[:n])
		err = binary.Read(buf, binary.BigEndian, hresp)
		if err != nil {
			return nil, err
		}
		if hresp.Action == ActionError {
			return nil, errors.New(buf.String())
		}
		err = binary.Read(buf, binary.BigEndian, bresp)
		if err != nil {
			return nil, err
		}
		if len(buf.Bytes()) > 0 {
			return buf, nil
		}
		break
	}
	return nil, nil
}

// Connect method connects to a tracker and does the handshake
func (tc *TrackerConnection) Connect() (err error) {
	//fmt.Println("connecting to", tc.URL.Host)
	tc.Socket, err = net.Dial("udp", tc.URL.Host)
	if err != nil {
		return err
	}

	transactionID := int32(rand.Uint32())
	err = tc.write(&HeaderRequest{
		ConnectionID:  connectionIDInitial,
		Action:        ActionConnect,
		TransactionID: transactionID,
	}, nil)
	if err != nil {
		return err
	}

	var cbresp ConnectingBodyResponse
	var hresp HeaderResponse
	bytesLeft, err := tc.read(&hresp, &cbresp)
	if err != nil {
		return nil
	}
	if hresp.TransactionID != transactionID {
		return fmt.Errorf("mismatch on transactionId: sent %d received %d",
			transactionID,
			hresp.TransactionID)
	}
	if bytesLeft != nil && len(bytesLeft.Bytes()) > 0 {
		fmt.Println("some bytes left to read", bytesLeft.String())
	}
	tc.ConnectionID = cbresp.ConnectionID
	fmt.Println("new tracker:", tc.URL.Host)
	return nil
}

// Announce announces current client to Tracker
func (tc *TrackerConnection) Announce() (err error) {
	key := rand.Uint32()
	err = tc.write(
		&HeaderRequest{
			ConnectionID:  tc.ConnectionID,
			Action:        ActionAnnounce,
			TransactionID: tc.TransactionID,
		},
		&AnnouncingBodyRequest{
			InfoHash:   tc.InfoHash,
			PeerID:     tc.PeerID,
			Downloaded: tc.Downloaded,
			Uploaded:   tc.Uploaded,
			Event:      EventStarted,
			IP:         0,
			Key:        key,
			NumWant:    -1,
			Port:       tc.ListenPort,
			Extensions: 0,
		})
	if err != nil {
		return err
	}
	var hresp HeaderResponse
	var bresp AnnouncingBodyResponse
	peers, err := tc.read(&hresp, &bresp)
	if err != nil {
		return err
	}

	for {
		var pr PeerResponse
		err = binary.Read(peers, binary.BigEndian, &pr)
		if err != nil {
			return err
		}
		tc.Peers = append(tc.Peers, pr)
		if len(peers.Bytes()) > 0 {
			continue
		}
		break
	}
	return nil
}
