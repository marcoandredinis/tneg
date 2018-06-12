package metadata

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// MagnetURIXT is part of MagnetURI Struct (XT field)
type MagnetURIXT struct {
	Position  int
	HashType  []string
	HashValue string
}

// MagnetURI represents a MagnetURI link
type MagnetURI struct {
	// https://en.wikipedia.org/wiki/Magnet_URI_scheme
	DN []string      // Display Name - Filename
	XL []string      // Exact Length - Size in bytes
	XT []MagnetURIXT // Exact Topic - URN containing hash type and file hash
	AS []string      // Acceptable Source - Web link to the file online
	XS []string      // Exact Source - P2P link
	KT []string      // Keyword Topic - Key words for search
	MT []string      // Manifest Topic - Link to the metafile that contains a list of magneto (MAGMA - MAGnet MAnifest)
	TR []string      // Address Tracker - Tracker URL for BitTorrent downloads
	WS []string      // Web Seeds: https://wiki.theory.org/BitTorrent_Magnet-URI_Webseeding
}

func parseXT(xt string, p int) (mx MagnetURIXT, err error) {
	s := strings.Split(xt, ":")
	if len(s) < 3 || s[0] != "urn" {
		return MagnetURIXT{}, fmt.Errorf("Invalid XT parameter: %q", xt)
	}
	hType := s[1 : len(s)-1] // urn:tree:tiger:<hash> => returns [tree, tiger]
	hValue := s[len(s)-1]    // urn:sha1:<hash> => returns hash
	return MagnetURIXT{Position: p, HashType: hType, HashValue: hValue}, nil
}

// ParseMagnetURI receives a string and returns a MagnetURI struct
func ParseMagnetURI(uri string) (m MagnetURI, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return MagnetURI{}, fmt.Errorf("ParseMagnetURI.Parse: %s", err)
	}

	if u.Scheme != "magnet" {
		return MagnetURI{}, fmt.Errorf("Not a magneturi scheme: %q", u.Scheme)
	}

	for k, v := range u.Query() {
		if k == "dn" {
			m.DN = v
		} else if k == "xl" {
			m.XL = v
		} else if k == "as" {
			m.AS = v
		} else if k == "xs" {
			m.XS = v
		} else if k == "kt" {
			m.KT = v
		} else if k == "mt" {
			m.MT = v
		} else if k == "tr" {
			m.TR = v
		} else if k == "xt" {
			// xt=urn:sha1:<HASH>
			for _, x := range v {
				parsedXT, err := parseXT(x, 0)
				if err != nil {
					return MagnetURI{}, err
				}
				m.XT = append(m.XT, parsedXT)
			}
		} else if strings.HasPrefix(k, "xt.") {
			// xt.1=urn:sha1:<HASH>&xt.2=urn:sha1:<OTHERHASH>
			s := strings.Split(k, ".")
			if len(s) != 2 {
				return MagnetURI{}, fmt.Errorf("Invalid XT index: %q", k)
			}
			i, err := strconv.Atoi(s[1])
			if err != nil {
				return MagnetURI{}, fmt.Errorf("Invalid XT index value: %q", k)
			}
			for _, x := range v {
				parsedXT, err := parseXT(x, i)
				if err != nil {
					return MagnetURI{}, err
				}
				m.XT = append(m.XT, parsedXT)
			}
		}
	}
	return m, nil
}
