package metadata

import (
	"encoding/hex"
	"fmt"
	"net/url"
)

// Torrent struct is the metadata needed to download a torrent
type Torrent struct {
	HashValue   [20]byte
	Trackers    []*url.URL
	DisplayName string
}

// TorrentFromMagnetURI creates a torrent from a MagnetURI
func TorrentFromMagnetURI(m MagnetURI) (t Torrent, err error) {
	if len(m.DN) == 0 || m.DN[0] == "" {
		return Torrent{}, fmt.Errorf("MagnetURI does not contain DN parameter (Display Name)")
	}
	t.DisplayName = m.DN[0]

	if len(m.TR) == 0 {
		return Torrent{}, fmt.Errorf("MagnetURI does not contain TR parameter (Trackers)")
	}
	for _, e := range m.TR {
		trackerURL, err := url.Parse(e)
		if err != nil {
			return Torrent{}, err
		}
		t.Trackers = append(t.Trackers, trackerURL)
	}

	// we do not recognize multiple XT parameters for Torrent URIs
	if len(m.XT) != 1 {
		return Torrent{}, fmt.Errorf("magnetURI does not contain a single XT parameter (Exact Topic) %+v", m.XT)
	}
	xt := m.XT[0]
	if len(xt.HashType) != 1 { // our HashType must be ["btih"]
		return Torrent{}, fmt.Errorf("magnetURI: XT parameter does not contain HashType value")
	}
	if xt.HashType[0] != "btih" {
		return Torrent{}, fmt.Errorf("MagnetURI: HashType of XT param: expected btih, got: %s", xt.HashType[0])
	}

	if len(xt.HashValue) != 40 {
		return Torrent{}, fmt.Errorf("MagnetURI: HashValue length must be 40, got %d", len(xt.HashValue))
	}
	byteArray := make([]byte, hex.DecodedLen(len(xt.HashValue)))
	n, err := hex.Decode(byteArray, []byte(xt.HashValue))
	if err != nil {
		return Torrent{}, err
	}
	if n != 20 {
		return Torrent{}, fmt.Errorf("MagnetURI: HashValue length must be 40, got %d", len(xt.HashValue))
	}
	copy(t.HashValue[:], byteArray)
	return t, nil
}
