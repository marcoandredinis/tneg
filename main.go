package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/marcoandredinis/tneg/client"
	"github.com/marcoandredinis/tneg/metadata"
)

func testParseMagnet() {
	f, err := os.Open("./test_assets/magnet_links.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			_ = fmt.Errorf("error closing file %+v", f)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		m, err := metadata.ParseMagnetURI(line)
		if err != nil {
			_ = fmt.Errorf("error processing URI: %s", line)
		}

		t, err := metadata.TorrentFromMagnetURI(m)
		if err != nil {
			_ = fmt.Errorf("error processing torrent from MagnetURI")
		}

		fmt.Println(t)
	}
}

func testClient() {
	s := "magnet:?xt=urn:btih:655676fc05fb5b56e9b5553f8b6c8be81e5ba29f&dn=ptwiki-20170205.slob&tr=udp%3A%2F%2Fopen.demonii.com%3A1337&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.ccc.de%3A80"
	m, err := metadata.ParseMagnetURI(s)
	if err != nil {

		_ = fmt.Errorf("error processing URI")
	}
	t, err := metadata.TorrentFromMagnetURI(m)
	if err != nil {
		_ = fmt.Errorf("error processing torrent from MagnetURI")
	}

	c, err := client.FromTorrent(t)
	if err != nil {
		_ = fmt.Errorf("error processing Torrent: %s", t)
	}

	fmt.Println(c)
	err = c.StartTorrent()
	if err != nil {
		_ = fmt.Errorf("error starting torrent")
	}
}

func main() {
	testParseMagnet()
	testClient()
}
