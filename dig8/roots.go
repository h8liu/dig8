package dig8

import (
	"fmt"
	"net"
)

// MakeRoots makes the name server set for the root.
func MakeRoots() *ZoneServers {
	ret := NewZoneServers(D("."))

	ns := func(n, ip string) {
		ret.Add(
			D(fmt.Sprintf("%s.root-servers.net", n)),
			net.ParseIP(ip),
		)
	}

	// see en.wikipedia.org/wiki/Root_name_server for reference
	// (last update: year 2012)
	ns("a", "198.41.0.4")     // Verisign
	ns("b", "192.228.79.201") // USC-ISI
	ns("c", "192.33.4.12")    // Cogent
	ns("d", "128.8.10.90")    // U Maryland
	ns("e", "192.203.230.10") // NASA
	ns("f", "192.5.5.241")    // Internet Systems Consortium
	ns("g", "192.112.36.4")   // DISA
	ns("h", "128.63.2.53")    // U.S. Army Research Lab
	ns("i", "192.36.148.17")  // Netnod
	ns("j", "198.41.0.10")    // Verisign
	ns("k", "193.0.14.129")   // RIPE NCC
	ns("l", "199.7.83.42")    // ICANN
	ns("m", "202.12.27.33")   // WIDE Project

	return ret
}
