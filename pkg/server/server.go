package server

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	MaxPacketSizeInBytes int = 512    // max packet server will accept. Will do for now.
)

var clientDetails ClientDetails
var responsePacket []byte

// generatePacket just generates a byte array of 1's.
// Can replace this with something more sensible later?
func generatePacket( packetSize int) []byte {
	packet := make([]byte, packetSize)
	for i:=0;i<packetSize;i++ {
		packet[i] = 1
	}
	return packet
}

// constantly loop and respond to any client IPs coming in.
func respondToClient(ch chan ClientResponse ) {
	for {
		resp := <- ch
		conn, ok := clientDetails.ClientConnMap[ resp.Addr]

		if !ok {
			// create connection...
			// made up port.... not expecting client to be listening anyway
			c, err := net.Dial("udp", resp.Addr+":5000")
			if err != nil {
				fmt.Printf("ERROR!!! : Unable to reply to %s\n", resp.Addr)
			}
			clientDetails.AddClient(resp.Addr, c)
			conn = c
		}

		fmt.Printf("remote addr %s\n", conn.RemoteAddr().String())
		packet := generatePacket( resp.PacketSize)
		// then send traffic
		_, err := conn.Write(packet)
		if err != nil {
			fmt.Printf("ERROR!!! : Unable to write to %s\n", resp.Addr)
		}
	}
}

func RunServer(port int, verbose bool, serverRespond bool) {

	fmt.Printf("running server on port %d\n", port)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{ Port: port})
	if err != nil {
		log.Fatalf("Unable to listen %s\n", err.Error())
	}
	defer conn.Close()

	// request coming in. Allocated bytes up front.
	// reading request from main thread.... will this be a bottleneck?
	requestBuffer := make([]byte, MaxPacketSizeInBytes)

  byteCount := 0
  lastCheckedByteCount := 0

	// setup ticker to count bytes.
	ticker := time.NewTicker(1 * time.Second)

  // ticker to display received bitrate.
  if verbose {
	  go func() {
		  for _ = range ticker.C {
			  if lastCheckedByteCount == 0 {
				  lastCheckedByteCount = byteCount
			  } else {
				  diffByteCount := byteCount - lastCheckedByteCount
				  lastCheckedByteCount = byteCount

				  fmt.Printf("rate is %d KB per second\n", (diffByteCount / 1024))
			  }
		  }
	  }()
  }

	// channel to indicate traffic.
	// buffer it... but hopefully wont fill buffer at all.
	ch := make(chan ClientResponse, 10000)

	// store client cache for replying.
	clientDetails = NewClientDetails()
	if serverRespond {
		go respondToClient(ch)
	}

	fmt.Printf("start listening\n")
	for {
		n, clientAddr, err := conn.ReadFromUDP(requestBuffer)
		if err != nil {
			log.Printf("Error while reading UDP packet. Probably due to number of bytes in UDP packet great than allowed max of %d\n", MaxPacketSizeInBytes)
			log.Printf("error with client addr %s\n",  err.Error())
			continue
		}
		byteCount += n

		if serverRespond {
			ch <- ClientResponse{clientAddr.IP.String(), n}
		}
		//log.Printf("client addr %s : %d\n", clientAddr.IP.String(), clientAddr.Port)
	}

	ticker.Stop()
}
