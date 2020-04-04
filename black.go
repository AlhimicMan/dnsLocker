package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type BlackList struct {
	data map[string]struct{}
}

func NewBlackList() *BlackList {
	return &BlackList{
		data: make(map[string]struct{}),
	}
}

func (b *BlackList) Add(server string) bool {
	server = strings.Trim(server, " ")
	if len(server) == 0 {
		return false
	}

	if !strings.HasSuffix(server, ".") {
		server += "."
	}
	b.data[server] = struct{}{}

	return true
}

func (b *BlackList) AddList(servers []string) (count int) {
	for _, server := range servers {
		if b.Add(server) {
			count++
		}
	}

	return
}

func (b *BlackList) Contains(server string) bool {
	for regH := range b.data {
		match, err := regexp.Match(regH, []byte(server))
		if match && err == nil {
			return true
		}
	}
	return false
}

func UpdateList() *BlackList {
	list := NewBlackList()

	for _, v := range config.Blocklist {
		resp, err := http.Get(v)
		if err != nil {
			log.Println("[black] Can't load", v)
			continue
		}

		if resp.StatusCode != 200 {
			log.Println("[black] Status code of", v, "!= 200")
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("[black] Can't read body of", v)
			continue
		}

		r := regexp.MustCompile("server=/(.*?)/")
		data2 := r.ReplaceAllString(string(data), "$1")
		data2 = strings.Replace(data2, "\r", "", -1)

		servers := strings.Split(data2, "\n")
		cnt := list.AddList(servers)
		log.Println("[black] Loaded", cnt, "servers from", v)
	}

	for _, blockDomain := range config.BlockHosts {
		runeStr := []rune(blockDomain)
		if runeStr[0] == rune('(') {
			runeStr = runeStr[1:]
		}
		if runeStr[len(runeStr)-1] == rune(')') {
			runeStr = runeStr[:len(runeStr)-2]
		}
		blockDomain = string(runeStr)
		list.data[blockDomain] = struct{}{}
	}

	return list
}

func listUpdater(configChanged chan struct{}) {
	for {
		<-configChanged
		// time.Sleep(config.UpdateInterval)
		blackList = UpdateList()
		log.Println("BlockList updated")
	}
}
