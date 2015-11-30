package driver

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	netApi "github.com/docker/libnetwork/drivers/remote/api"
	ipamApi "github.com/docker/libnetwork/ipams/remote/api"
	"github.com/jc-m/test-docker-plugin/routed/server"
	"github.com/vishvananda/netlink"
	"math/rand"
	"net"
	"time"
)

type routedEndpoint struct {
	iface         string
	macAddress    net.HardwareAddr
	hostInterface string
	ipv4Address   *net.IPNet
}

type routedNetwork struct {
	id        string
	endpoints map[string]*routedEndpoint
}

type routedPool struct {
	id           string
	subnet       *net.IPNet
	gateway      *net.IPNet
	allocatedIPs map[string]bool
}

type driver struct {
	version string
	network *routedNetwork
	pool    *routedPool
	mtu     int
}

func New(version string) (server.Driver, error) {
	network, _ := netlink.ParseIPNet("100.64.0.0/10")
	gateway, _ := netlink.ParseIPNet("100.64.0.1/32")
	pool := &routedPool{
		id:           "myPool",
		subnet:       network,
		allocatedIPs: make(map[string]bool),
		gateway:      gateway,
	}
	rnet := &routedNetwork{
		endpoints: make(map[string]*routedEndpoint),
	}
	pool.allocatedIPs[fmt.Sprintf("%s", gateway)] = true
	return &driver{
		version: version,
		pool:    pool,
		network: rnet,
	}, nil
}

// ======= Driver functions

func (driver *driver) GetCapabilities() (*netApi.GetCapabilityResponse, error) {
	caps := &netApi.GetCapabilityResponse{
		Scope: "local",
	}
	log.Debugf("Get capabilities: responded with %+v", caps)
	return caps, nil
}

func (driver *driver) CreateNetwork(create *netApi.CreateNetworkRequest) error {
	log.Debugf("Create network request %+v", create)

	driver.network = &routedNetwork{id: create.NetworkID, endpoints: make(map[string]*routedEndpoint)}
	log.Infof("Create network %s", create.NetworkID)

	return nil
}

func (driver *driver) DeleteNetwork(delete *netApi.DeleteNetworkRequest) error {
	log.Debugf("Delete network request: %+v", delete)
	driver.network = nil
	log.Infof("Destroying network %s", delete.NetworkID)
	return nil
}

func (driver *driver) CreateEndpoint(create *netApi.CreateEndpointRequest) (*netApi.CreateEndpointResponse, error) {
	log.Debugf("Create endpoint request %+v", create)
	endID := create.EndpointID
	reqIface := create.Interface
	log.Debugf("Requested Interface %+v", reqIface)
	addr, _ := netlink.ParseIPNet(reqIface.Address)
	ep := &routedEndpoint{
		ipv4Address: addr,
	}
	driver.network.endpoints[endID] = ep

	hw := make(net.HardwareAddr, 6)
	hw[0] = 0xde
	hw[1] = 0xad
	copy(hw[2:], addr.IP.To4())

	respIface := &netApi.EndpointInterface{
		MacAddress: hw.String(),
	}
	resp := &netApi.CreateEndpointResponse{
		Interface: respIface,
	}

	log.Infof("Creating endpoint %s %+v", endID, resp)
	return resp, nil
}

func (driver *driver) DeleteEndpoint(delete *netApi.DeleteEndpointRequest) error {
	log.Debugf("Delete endpoint request: %+v", delete)

	log.Infof("Deleting endpoint %s", delete.EndpointID)
	return nil
}

func (driver *driver) EndpointInfo(req *netApi.EndpointInfoRequest) (*netApi.EndpointInfoResponse, error) {
	log.Debugf("Endpoint info request: %+v", req)

	log.Infof("Endpoint info %s", req.EndpointID)
	return &netApi.EndpointInfoResponse{Value: map[string]interface{}{}}, nil
}

func (driver *driver) JoinEndpoint(j *netApi.JoinRequest) (*netApi.JoinResponse, error) {
	log.Debugf("Join endpoint request: %+v", j)
	log.Debugf("Joining endpoint %s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey)

	tempName := j.EndpointID[:4]
	hostName := "vethr" + j.EndpointID[:4]

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:   hostName,
			TxQLen: 0,
		},
		PeerName: tempName,
	}
	log.Infof("Adding link %+v", veth)
	if err := netlink.LinkAdd(veth); err != nil {
		log.Errorf("Unable to add link %+v:%+v", veth, err)
		return nil, err
	}
	if err := netlink.LinkSetMTU(veth, 1500); err != nil {
		log.Errorf("Error setting the MTU %s", err)
	}
	log.Infof("Bringing link up %+v", veth)
	if err := netlink.LinkSetUp(veth); err != nil {
		log.Errorf("Unable to bring up %+v: %+v", veth, err)
		return nil, err
	}
	ep := driver.network.endpoints[j.EndpointID]
	ep.iface = hostName

	iface, _ := netlink.LinkByName(hostName)

	route := netlink.Route{
		LinkIndex: iface.Attrs().Index,
		Dst:       ep.ipv4Address,
	}

	log.Infof("Adding route %+v", route)
	if err := netlink.RouteAdd(&route); err != nil {
		log.Errorf("Unable to add route %+v: %+v", route, err)
	}

	respIface := &netApi.InterfaceName{
		SrcName:   tempName,
		DstPrefix: "eth",
	}
	sandboxRoute := netApi.StaticRoute{
		Destination: "0.0.0.0/0",
		RouteType:   1, // CONNECTED
		NextHop:     "",
	}
	resp := &netApi.JoinResponse{
		InterfaceName: respIface,
		StaticRoutes:  []netApi.StaticRoute{sandboxRoute},
	}
	log.Infof("Join Request Response %+v", resp)

	return resp, nil
}

func (driver *driver) LeaveEndpoint(leave *netApi.LeaveRequest) error {
	log.Debugf("Leave request: %+v", leave)
	ep := driver.network.endpoints[leave.EndpointID]
	link, err := netlink.LinkByName(ep.iface)
	if err == nil {
		log.Debugf("Deleting host interface %s", ep.iface)
		netlink.LinkDel(link)
	} else {
		log.Debugf("interface %s not found", ep.iface)
	}
	log.Infof("Leaving %s:%s", leave.NetworkID, leave.EndpointID)
	return nil
}

func (driver *driver) GetDefaultAddressSpaces() (*ipamApi.GetAddressSpacesResponse, error) {
	spaces := &ipamApi.GetAddressSpacesResponse{
		LocalDefaultAddressSpace:  "Testlocal",
		GlobalDefaultAddressSpace: "TestRemote",
	}
	log.Infof("Get default addresse spaces: responded with %+v", spaces)
	return spaces, nil
}

/// IPAM driver

func (driver *driver) RequestPool(p *ipamApi.RequestPoolRequest) (*ipamApi.RequestPoolResponse, error) {
	log.Debugf("Pool Request request: %+v", p)

	cidr := fmt.Sprintf("%s", driver.pool.subnet)
	id := driver.pool.id
	gateway := fmt.Sprintf("%s", driver.pool.gateway)
	pool := &ipamApi.RequestPoolResponse{
		PoolID: id,
		Pool:   cidr,
		Data:   map[string]string{"com.docker.network.gateway": gateway},
	}

	log.Infof("Pool Request: responded with %+v", pool)
	return pool, nil
}

func (driver *driver) RequestAddress(a *ipamApi.RequestAddressRequest) (*ipamApi.RequestAddressResponse, error) {
	log.Debugf("Address Request request: %+v", a)

again:
	// just generate a random address
	rand.Seed(time.Now().UnixNano())
	ip := driver.pool.subnet.IP.To4()
	ip[3] = byte(rand.Intn(254))
	netIP := fmt.Sprintf("%s/32", ip)
	log.Infof("ip:%s", netIP)

	_, ok := driver.pool.allocatedIPs[netIP]

	if ok {
		goto again
	}
	driver.pool.allocatedIPs[netIP] = true
	resp := &ipamApi.RequestAddressResponse{
		Address: fmt.Sprintf("%s", netIP),
	}

	log.Infof("Addresse request response: %+v", resp)
	return resp, nil
}

func (driver *driver) ReleaseAddress(a *ipamApi.ReleaseAddressRequest) error {
	log.Debugf("Address Release request: %+v", a)
	ip := fmt.Sprintf("%s/32", a.Address)

	delete(driver.pool.allocatedIPs, ip)

	log.Infof("Addresse release %s from %s", a.Address, a.PoolID)
	return nil
}

func (driver *driver) ReleasePool(p *ipamApi.ReleasePoolRequest) error {
	log.Debugf("Pool Release request: %+v", p)

	log.Infof("Pool release %s ", p.PoolID)
	return nil
}
