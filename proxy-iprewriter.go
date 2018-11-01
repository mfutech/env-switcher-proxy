package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/elazarl/goproxy.v1"
)

type HostRewriter struct {
	Verbose      bool
	RerouteTable map[string]string
}

var rerouteTable = make(map[string]string)

/*
var rerouteTable = map[string]string {

	"www.tribufufu.net:80":  "www.google.ch:80",
	"www.tribufufu.net:443": "www.google.ch:443",
}
*/
/*
func (hrw *HostRewriter) newHostRewriter() {
	hrw.RerouteTable = make(map[string]string)
}
*/
// RewriteAddr use reroute table to translate address
func (hrw *HostRewriter) RewriteAddr(network, addr string) (rNetwork, rAddr string) {
	//fmt.Printf(">> RewriteAddr (%s %s) >>\n", network, addr)
	rNetwork, rAddr = network, addr
	newAddr, ok := hrw.RerouteTable[addr]
	if ok {
		rAddr = newAddr
		if hrw.Verbose {
			fmt.Printf("reroute match %s -> %s\n", addr, rAddr)
		}
	}
	//fmt.Printf("<< RewriteAddr (%s %s) <<\n", rNetwork, rAddr)
	return
}

func main() {
	var hrw HostRewriter

	// get falgs
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("addr", ":8080", "proxy listen address")
	configFilename := flag.String("config", "config", "configuration filename (without extension)")
	flag.Parse()

	// get configuration file
	configName := strings.Split(*configFilename, ".")[0] //we only take the "basename" part
	viper.SetConfigName(configName)                      // defining the name of the config (without extension)
	viper.AddConfigPath("conf")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("error reading config : %s", err))
	}

	// create host rewriter
	hrw.RerouteTable = viper.GetStringMapString("routingTable")
	hrw.Verbose = *verbose

	// creating a proxy
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose

	// intercept the establishment of configuration
	// and rewrite destination IP
	proxy.Tr.Dial = func(network, addr string) (c net.Conn, err error) {
		network, addr = hrw.RewriteAddr(network, addr)
		c, err = net.Dial(network, addr)
		return
	}

	// run the proxy
	log.Fatal(http.ListenAndServe(*addr, proxy))
}
