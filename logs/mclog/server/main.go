package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"

	_ "sirherobrine23.com.br/go-bds/go-bds/logs/bedrock"
	_ "sirherobrine23.com.br/go-bds/go-bds/logs/java"
	"sirherobrine23.com.br/go-bds/go-bds/logs/mclog"
)

var (
	PortPoint = flag.Int("port", 0, "Port to listen http server")
	Rootdir   = flag.String("root", filepath.Join(os.TempDir(), "bdsmclogs"), "Folder to save log files")
)

func main() {
	flag.Parse()
	rootDir, port := *Rootdir, uint16(*PortPoint)
	os.MkdirAll(rootDir, 0755)

	limits := mclog.Limits{}
	handler := mclog.NewHandler(limits, mclog.Local(rootDir))

	listen, err := net.ListenTCP("tcp", net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.IPv4Unspecified(), port)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot listen http server: %s\n", err)
		os.Exit(1)
		return
	}
	fmt.Fprintf(os.Stderr, "HTTP server listening on %s\n", listen.Addr().String())
	http.Serve(listen, handler)
}
