package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// SourceCounter keeps track of the number of received messages from each
// source until the next Report (at which point, the counters are cleared)
type SourceCounter struct {
	list map[string]*int64
	mu   sync.Mutex
}

// NewSourceCounter returns a new instantiated SourceCounter
func NewSourceCounter() *SourceCounter {
	return &SourceCounter{
		list: make(map[string]*int64),
	}
}

func (s *SourceCounter) Add(src string) {
	cnt, ok := s.list[src]
	if !ok {
		s.mu.Lock()
		cnt = new(int64)
		s.list[src] = cnt
		s.mu.Unlock()
	}
	*cnt = *cnt + 1
}

func (s *SourceCounter) Report() {
	s.mu.Lock()

	for ip, count := range s.list {
		log.Printf("%s: %d", ip, *count)
	}

	s.list = make(map[string]*int64)
	s.mu.Unlock()
}

func main() {
	var err error

	ip := "::"
	if len(os.Args) > 1 {
		ip = os.Args[1]
	}

	port := 10100
	if len(os.Args) > 2 {
		port, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalln("failed to parse port as a number:", err)
		}
	}

	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}

	b := make([]byte, 2048)

	sc := NewSourceCounter()
	go reporter(sc)
	go inputReporter(sc)

	for {
		cc, remote, err := conn.ReadFromUDP(b)
		if err != nil {
			fmt.Printf("net.ReadFromUDP() error: %s\n", err)
		}

		sc.Add(remote.String())

		_, err = conn.WriteTo(b[0:cc], remote)
		if err != nil {
			fmt.Printf("net.WriteTo() error: %s\n", err)
		}
	}
}

func reporter(sc *SourceCounter) {
	t := time.NewTicker(5 * time.Minute)

	for {
		<-t.C
		sc.Report()
	}
}

func inputReporter(sc *SourceCounter) {
	r := bufio.NewReader(os.Stdin)
	for {
		_, _ = r.ReadString('\n')
		sc.Report()
	}
}
