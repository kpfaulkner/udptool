package main

import (
	"flag"
	"fmt"
	"github.com/kpfaulkner/udptool/pkg/client"
	"github.com/kpfaulkner/udptool/pkg/server"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func displayHelp() {
  fmt.Printf("UDPTool\n\n")
  fmt.Printf("Can run as server or client to help test throughput of UDP connections\n")
  fmt.Printf("args:\n")
  fmt.Printf("  -server : indicates if should be run as server. Default is run as a client\n")
	fmt.Printf("  -port : when running as server, which port should UDP server listen on\n")
	fmt.Printf("  -noclients : when running as a client, how many 'clients' should be simulated. These run concurrently\n")
	fmt.Printf("  -seconds : how long should clients generate traffic for\n")
	fmt.Printf("  -bps : how many BITS per second should each client send to the server\n")
	fmt.Printf("  -host : when running as client, what is the 'host:port' of the server\n")
	fmt.Printf("  -minclientaddr : min client address to use. Could be first of VPN connections (if VPN used)\n")
	fmt.Printf("  -noclientaddr : Number of client IP addresses to use. Assumed to be INCREMENTED from minClientAddress\n")
	fmt.Printf("  -serverrespond : Server should reply to client. Client wont actually listen, but it WILL generate traffic\n")

	fmt.Printf("  -packetsize : UDP packet size\n")
	fmt.Printf("  -verbose : prepare for spam...\n")

	fmt.Printf("  -help : this :)\n")
}

// new entry point for both server and client.
func main() {
  isServer := flag.Bool("server", false, "run tool as server. Default is as client")
	port := flag.Int("port", 5001, "run server on this port")
	noClients := flag.Int("noclients", 1, "number of concurrent clients to simulate")
	noSeconds := flag.Int("seconds", 10, "number of seconds test runs for")
	bps := flag.Int("bps", 68000, "number of BITS per second")
	packetSize := flag.Int("packetsize", 160, "UDP packet size in bytes")
	host := flag.String("host", "", "tells client which host:port the server is on")
	//clientAddr := flag.String("clientaddr", "127.0.0.1", "which client interface/ip to use. Defaults to 127.0.0.1")
  verbose := flag.Bool("verbose", false, "spamageddon")
	help := flag.Bool("help", false, "help")
	minClientAddress := flag.String("minclientaddr", "127.0.0.1", "min client address to use. Could be first of VPN connections (if VPN used)")
	noClientAddresses := flag.Int("noclientaddr", 1, "Number of client IP addresses to use. Assumed to be INCREMENTED from minClientAddress")
	serverRespond := flag.Bool("serverrespond", false, "Server should reply to client. Client wont actually listen, but it WILL generate traffic")

	profile := flag.Bool("profile", false, "enable remote profiling")


	flag.Parse()

	if *help {
		displayHelp()
		return
	}

	if !(*isServer) && (*host) == "" {
		fmt.Printf("No server host specified. Do not know where to send the data!\n\n")
		displayHelp()
		return
	}

	if (*profile) {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if *isServer {
		server.RunServer(*port, *verbose, *serverRespond)
	} else {
		client.RunClient(int32(*noClients), *noSeconds, *bps, *host, *verbose, *packetSize, *minClientAddress, *noClientAddresses)
	}

}


