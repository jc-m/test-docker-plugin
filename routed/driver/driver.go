package driver

import (
	"net"
	"math/rand"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
    netApi"github.com/docker/libnetwork/drivers/remote/api"
	ipamApi"github.com/docker/libnetwork/ipams/remote/api"
    "github.com/jc-m/test-docker-plugin/routed/server"
	"github.com/vishvananda/netlink"
)

const (
	networkType = "routed"
	ifaceID     = 1
	VethPrefix  = "vethr" 
)

type routedEndpoint struct {
	id              string
	iface           string
	macAddress      net.HardwareAddr
	hostInterface   string
	ipv4Addresses   []netlink.Addr
}

type routedNetwork struct {
	id        string
	endpoints map[string]*routedEndpoint
}

type driver struct{
	version	string
	network *routedNetwork
	mtu     int
}

func New(version string) (server.Driver, error) {
	
	// TODO Clean interface list
	
	return &driver{
		version:    version,
	}, nil
}

// ======= Driver functions

func (driver *driver) GetCapabilities() (*netApi.GetCapabilityResponse, error) {
	caps := &netApi.GetCapabilityResponse{
		Scope: "local",
	}	
	log.Infof("Get capabilities: responded with %+v", caps)
	return caps, nil
}

func (driver *driver) CreateNetwork(create *netApi.CreateNetworkRequest) error {
	log.Infof("Create network request %+v", create)
	// TODO Seems like there is more to do...
	
	driver.network = &routedNetwork{id: create.NetworkID, endpoints: make(map[string]*routedEndpoint)}
	log.Infof("Create network %s", create.NetworkID)
	
	return nil
}

func (driver *driver) DeleteNetwork(delete *netApi.DeleteNetworkRequest) error {
	log.Infof("Delete network request: %+v", delete)
	driver.network = nil
	log.Infof("Destroy network %s", delete.NetworkID)
	return nil
}

func (driver *driver) CreateEndpoint(create *netApi.CreateEndpointRequest) (*netApi.CreateEndpointResponse, error) {
	log.Infof("Create endpoint request %+v", create)
	endID := create.EndpointID
	reqIface := create.Interface
	log.Infof("Requested Interface %+v", reqIface)
	
//	respIface := &netApi.EndpointInterface{
//		Address: "172.18.1.2/32",
//		MacAddress: "EE:EE:EE:EE:EE:EE",
//	}
	respIface := &netApi.EndpointInterface{}
	// TODO do something
	resp := &netApi.CreateEndpointResponse{
		Interface: respIface,
	}

	log.Infof("Create endpoint %s %+v", endID, resp)
	return resp, nil
}

func (driver *driver) DeleteEndpoint(delete *netApi.DeleteEndpointRequest) error {
	log.Infof("Delete endpoint request: %+v", delete)
	// TODO fill the blank
	log.Infof("Delete endpoint %s", delete.EndpointID)
	return nil
}

func (driver *driver) EndpointInfo(req *netApi.EndpointInfoRequest) (*netApi.EndpointInfoResponse, error) {
	log.Infof("Endpoint info request: %+v", req)
	log.Infof("Endpoint info %s", req.EndpointID)
	return &netApi.EndpointInfoResponse{Value: map[string]interface{}{}}, nil
}

func (driver *driver) JoinEndpoint(j *netApi.JoinRequest) (*netApi.JoinResponse, error) {
	log.Infof("Delete endpoint request: %+v", &j)
	
	log.Infof("Join endpoint %s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey)
	return nil, nil
}

func (driver *driver) LeaveEndpoint(leave *netApi.LeaveRequest) error {
	log.Infof("Leave request: %+v", &leave)

	log.Infof("Leave %s:%s", leave.NetworkID, leave.EndpointID)
	return nil
}

func (driver *driver) GetDefaultAddressSpaces() (*ipamApi.GetAddressSpacesResponse, error) {
	spaces := &ipamApi.GetAddressSpacesResponse{
		LocalDefaultAddressSpace: "Testlocal",
		GlobalDefaultAddressSpace: "TestRemote",
	}	
	log.Infof("Get default addresse spaces: responded with %+v", spaces)
	return spaces, nil
}

func (driver *driver) RequestPool(p *ipamApi.RequestPoolRequest) (*ipamApi.RequestPoolResponse, error) {
	log.Infof("Pool Request request: %+v", p)
	
	pool := &ipamApi.RequestPoolResponse{
		PoolID: "myPool",
		Pool: "100.64.0.0/10",
	}	
	log.Infof("Pool Request: responded with %+v", pool)
	return pool, nil
}

func (driver *driver) RequestAddress(a *ipamApi.RequestAddressRequest) (*ipamApi.RequestAddressResponse, error) {
	log.Infof("Address Request request: %+v", a)
	// just generate a random address
	rand.Seed(time.Now().UnixNano())
	address := &ipamApi.RequestAddressResponse{
		Address: fmt.Sprintf("100.64.0.%d/32",rand.Intn(254)),
	}	
	log.Infof("addresse request: responded with %+v", address)
	return address, nil
}

func (driver *driver) ReleaseAddress(a *ipamApi.ReleaseAddressRequest) error {
	log.Infof("Address Release request: %+v", a)

	log.Infof("addresse release %s from %s", a.Address, a.PoolID)
	return nil
}

func (driver *driver) ReleasePool(p *ipamApi.ReleasePoolRequest) error {
	log.Infof("Pool Release request: %+v", p)

	log.Infof("pool release %s ", p.PoolID)
	return nil
}