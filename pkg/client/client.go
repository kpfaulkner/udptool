package client

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	MaxPacketSizeInBytes int = 512    // max packet server will accept. Will do for now.
)

// used for synchronising the simulated clients.
var wg sync.WaitGroup
var triggerWg sync.WaitGroup
var clientsRun int32



// generatePacket just generates a byte array of 1's.
// Can replace this with something more sensible later?
func generatePacket( packetSize int) []byte {
	packet := make([]byte, packetSize)
	for i:=0;i<packetSize;i++ {
		packet[i] = 1
	}
	return packet
}

// sendPacketsOfSpecificSize sends out a series of UDP packets of a given size up to "totalBitsToSend"
// All of these will need to go out within a second, otherwise we're not going to be able to emulate a RPS
// type of situation.
func sendPacketsOfSpecificSize( identifier int32, packetSizeInBytes int, totalBitsToSend int, conn net.Conn ) error {

	noPackets := (totalBitsToSend/8) / packetSizeInBytes
	packet := generatePacket(packetSizeInBytes)

	for i:=0;i<noPackets;i++ {

		n, err := conn.Write(packet)
		if err != nil {
			log.Printf("Unable to write bytes!!! error! %s\n", err.Error())
			return err
		}

		if n != packetSizeInBytes {
			log.Printf("unable to send all of packet... something wrong\n")
		}
	}

	return nil
}

func createDialerForLocalAddress( addr string ) (*net.Dialer, error) {

	localAddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		fmt.Printf("Unable to resolve local address %s : %s\n", addr, err.Error())
		return nil, err
	}

	localUDPAddr := net.UDPAddr{
		IP: localAddr.IP,
	}

	dialer := net.Dialer{ LocalAddr: &localUDPAddr }
	return &dialer, nil
}

func connectAndSend(conn net.Conn, identifier int32, runForNSeconds int, bps int, verbose bool, packetSize int, clientAddr string) error {

	if verbose {
		fmt.Printf("Using address %s\n", clientAddr)
	}

	triggerWg.Wait()
	clientStartTime := time.Now()
	for i := 0; i< runForNSeconds; i++ {
		time1 := time.Now()
		sendPacketsOfSpecificSize(identifier, packetSize, bps, conn)
		time2 := time.Now()
		durInMS := time2.Sub(time1).Nanoseconds()/1000000

		if verbose {
			fmt.Printf("id %d : packet took %d ms\n", identifier, durInMS)
		}

		sleepDuration := 1000 - durInMS
		if sleepDuration > 0 {
			if verbose {
				fmt.Printf("sleep time %d ms\n", sleepDuration)
			}
			time.Sleep(  time.Duration(sleepDuration) * time.Millisecond)
		} else {
			// took longer than a second. Need to warn user....
			fmt.Printf("PERF ISSUE!!!! : Packet sent to id %d took longer than a second.", identifier)
		}
	}
	clientEndTime := time.Now()
	fmt.Printf("completed %d in %d seconds\n", identifier, int32(clientEndTime.Sub(clientStartTime).Seconds()))
	conn.Close()
	wg.Done()
	return nil
}

func nextIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3)
}

// just generate list of IPs we'll be using.
// precalculating is overkill, but easier for now.
func generateIPList( minClientAddr string,count int, verbose bool) []string {

	ipList := []string{}
	ip := net.ParseIP(minClientAddr)
	for i := 0;i<count;i++ {
		ip = ip.To4()
		ipList = append(ipList, ip.String())
		ip = nextIP(ip, 1)
	}

	if verbose {
		fmt.Printf("list length %d\n", len(ipList))
		for _,i := range ipList{
			fmt.Printf("ip %s\n", i)
		}

	}
	return ipList
}

func generateConn(clientAddr string, server string ) (*net.Conn, error) {
	dialer, err := createDialerForLocalAddress(clientAddr)
	if err != nil {
		fmt.Printf("unable to create dialer : error %s\n", err.Error())
		return nil, err
	}

	conn, err := dialer.Dial("udp", server)
	if err != nil {
		//fmt.Printf("connectAndSend error %s\n", err.Error())
		return nil, err
	}

  return &conn, nil
}
// run "noClients" number of clients across a certain number of IPs.
// They do not have to be equal, the clients will just loop over again.
func RunClient(noClients int32, noSeconds int, bps int,  server string, verbose bool, packetSize int, minClientAddr string, noIpAddresses int) {

	ipList := generateIPList(minClientAddr, noIpAddresses, verbose)

	clientsRun = 0
	ipCount := 0

	// use this as a trigger for all goroutines.
	// Have them wait on this...and will only be "done" once all goroutines are generated.
	triggerWg.Add(1)

	fmt.Printf("Total number of clients to generate %d\n", noClients)
	wg.Add(int(noClients))
	//for clientsRun < noClients {
	for c:=int32(0); c<noClients ; {
		cc := c
		if ipCount >= noIpAddresses {
			ipCount=0
		}
		ip := ipList[ipCount]

		conn, err := generateConn( ip, server)
		if err != nil {
			// skip this.... need another IP
			ipCount++
			continue
		}

		//go connectAndSend(cc, noSeconds, bps, ip, server, verbose, packetSize)
		go connectAndSend( *conn, cc, noSeconds, bps, verbose, packetSize, ip)

		time.Sleep(time.Duration(100) * time.Millisecond)
		ipCount++
		c++

	}

	triggerWg.Done()

	// wait for all clients to finish their work.
	wg.Wait()
}

