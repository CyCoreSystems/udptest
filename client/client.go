package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

var name string
var port int
var messageCount int

func init() {
	flag.StringVar(&name, "host", "", "host is the echo server to which we should connect")
	flag.IntVar(&port, "port", 10100, "port defines the UDP port to which we should connect")
	flag.IntVar(&messageCount, "count", 100, "count is the number of datagrams to send")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	if name == "" {
		log.Fatal("host is a required parameter")
	}

	nameport := name + ":" + strconv.Itoa(port)

	conn, err := net.Dial("udp", nameport)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Local address: %v\n", conn.LocalAddr())
	fmt.Printf("Remote address: %v\n", conn.RemoteAddr())

	b := []byte("abcdefghijklmnopqrstuvwxyz0123456789\n")

	go receiver(ctx, conn)

	time.Sleep(time.Second)

	for i := 0; i < messageCount; i++ {
		_, err = conn.Write(b)
		if err != nil {
			fmt.Printf("conn.Write() error: %s\n", err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(time.Second)
	cancel()

	fmt.Println("total sent messages:", messageCount)
	if err = conn.Close(); err != nil {
		time.Sleep(time.Second)
		log.Fatal(err)
	}
	time.Sleep(time.Second)
}

func receiver(ctx context.Context, conn io.Reader) {
	var cc int
	var count int
	var err error
	c := make([]byte, 40)

	for ctx.Err() == nil {
		if count == messageCount {
			break
		}
		cc, err = conn.Read(c)
		if err != nil {
			fmt.Printf("conn.Read() error: %s\n", err)
			break
		}
		if cc != 37 {
			fmt.Printf("ERROR: wrong bytes read: %d != %d", cc, 37)
		} else {
			count++
		}
	}
	fmt.Println("total read messages:", count)
}
