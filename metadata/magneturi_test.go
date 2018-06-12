package metadata

import (
	"testing"
)

func TestParseMagnetURIMultiXT(t *testing.T) {
	magnetURIString := "magneturi:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7"
	m, err := ParseMagnetURI(magnetURIString)
	if err != nil {
		t.Fatalf("Error: expected nil but got %s", err)
	}
	if len(m.XT) != 2 {
		t.Fatalf("Len(m.XT) expected %d but got %d", 2, len(m.XT))
	}
	if m.XT[0].Position != 1 {
		t.Fatalf("XT[0].Position expected %d but got %d", 1, m.XT[0].Position)
	}
	if len(m.XT[0].HashType) != 1 {
		t.Fatalf("Len(XT[0].HashType) expected %d but got %d", 1, len(m.XT[0].HashType))
	}
	if m.XT[0].HashType[0] != "sha1" {
		t.Fatalf("Array[0].HashType expected %s but got %s", "sha1", m.XT[0].HashType)
	}
	if m.XT[0].HashValue != "YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C" {
		t.Fatalf("Array[0].HashValue expected %s but got %s", "YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C", m.XT[0].HashValue)
	}
}

func TestParseMagnetURIMultiHashXT(t *testing.T) {
	magnetURIString := "magneturi:?xt=urn:ed2k:354B15E68FB8F36D7CD88FF94116CDC1&xt=urn:tree:tiger:7N5OAMRNGMSSEUE3ORHOKWN4WWIQ5X4EBOOTLJY&xt=urn:btih:QHQXPYWMACKDWKP47RRVIV7VOURXFE5Q&xl=10826029&dn=mediawiki-1.15.1.tar.gz&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&as=http%3A%2F%2Fdownload.wikimedia.org%2Fmediawiki%2F1.15%2Fmediawiki-1.15.1.tar.gz&xs=http%3A%2F%2Fcache.example.org%2FXRX2PEFXOOEJFRVUCX6HMZMKS5TWG4K5&xs=dchub://example.org"
	m, err := ParseMagnetURI(magnetURIString)
	if err != nil {
		t.Fatalf("Error: expected nil but got %s", err)
	}
	if len(m.XT) != 3 {
		t.Fatalf("Len(m.XT) expected %d but got %d", 3, len(m.XT))
	}
	if m.XT[0].Position != 0 {
		t.Fatalf("Array[0].Position expected %d but got %d", 0, m.XT[0].Position)
	}
	if len(m.XT[1].HashType) != 2 {
		t.Fatalf("Len(XT[1].HashType) expected %d but got %d", 2, len(m.XT[0].HashType))
	}
	if m.XT[0].HashType[0] != "ed2k" {
		t.Fatalf("Array[0].HashType expected %s but got %s", "ed2k", m.XT[0].HashType)
	}
	if m.XT[0].HashValue != "354B15E68FB8F36D7CD88FF94116CDC1" {
		t.Fatalf("Array[0].HashValue expected %s but got %s", "YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C", m.XT[0].HashValue)
	}
}
