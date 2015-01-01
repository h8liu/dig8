package dig8

import (
	"fmt"
	"net"
)

// NameServer saves a name server for a zone
type NameServer struct {
	Zone   *Domain
	Domain *Domain
	IP     net.IP
}

func (ns *NameServer) String() string {
	if ns.IP == nil {
		return fmt.Sprintf("%v ns %v",
			ns.Zone, ns.Domain,
		)
	}

	return fmt.Sprintf("%v ns %v(%v)",
		ns.Zone, ns.Domain, ns.IP,
	)
}

// Key returns the a string in ip@zone form.
func (ns *NameServer) Key() string {
	if ns.IP == nil {
		panic("unresolved")
	}

	return fmt.Sprintf("%v@%v", ns.IP, ns.Zone)
}
