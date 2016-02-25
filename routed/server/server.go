package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	log "github.com/Sirupsen/logrus"
    netApi"github.com/docker/libnetwork/drivers/remote/api"
	ipamApi"github.com/docker/libnetwork/ipams/remote/api"
	"github.com/gorilla/mux"
)


type Driver interface {
	GetCapabilities() (*netApi.GetCapabilityResponse, error)
	CreateNetwork(create *netApi.CreateNetworkRequest) error
	DeleteNetwork(delete *netApi.DeleteNetworkRequest) error
	CreateEndpoint(create *netApi.CreateEndpointRequest) (*netApi.CreateEndpointResponse, error)
	DeleteEndpoint(delete *netApi.DeleteEndpointRequest) error
	EndpointInfo(req *netApi.EndpointInfoRequest) (*netApi.EndpointInfoResponse, error)
	JoinEndpoint(j *netApi.JoinRequest) (response *netApi.JoinResponse, error error)
	LeaveEndpoint(leave *netApi.LeaveRequest) error
	GetIPAMCapabilities() (*ipamApi.GetCapabilityResponse, error)
	GetDefaultAddressSpaces() (*ipamApi.GetAddressSpacesResponse, error)
	RequestPool(p *ipamApi.RequestPoolRequest) (*ipamApi.RequestPoolResponse, error)
	RequestAddress(a *ipamApi.RequestAddressRequest) (*ipamApi.RequestAddressResponse, error)
	ReleaseAddress(a *ipamApi.ReleaseAddressRequest) error
	ReleasePool(a *ipamApi.ReleasePoolRequest) error
	
}

type server struct {
	d Driver
}



func Listen(socket net.Listener, driver Driver) error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(activate)

	server := &server{driver}

	// Network plugin methods
	router.Methods("POST").Path("/NetworkDriver.GetCapabilities").HandlerFunc(server.getCapabilities)
	router.Methods("POST").Path("/NetworkDriver.CreateNetwork").HandlerFunc(server.createNetwork)
	router.Methods("POST").Path("/NetworkDriver.DeleteNetwork").HandlerFunc(server.deleteNetwork)
	router.Methods("POST").Path("/NetworkDriver.CreateEndpoint").HandlerFunc(server.createEndpoint)
	router.Methods("POST").Path("/NetworkDriver.DeleteEndpoint").HandlerFunc(server.deleteEndpoint)
	router.Methods("POST").Path("/NetworkDriver.EndpointOperInfo").HandlerFunc(server.infoEndpoint)
	router.Methods("POST").Path("/NetworkDriver.Join").HandlerFunc(server.joinEndpoint)
	router.Methods("POST").Path("/NetworkDriver.Leave").HandlerFunc(server.leaveEndpoint)
	
	// IPAM plugin methods
	router.Methods("POST").Path("/IpamDriver.GetCapabilities").HandlerFunc(server.getIPAMCapabilities)
	router.Methods("POST").Path("/IpamDriver.GetDefaultAddressSpaces").HandlerFunc(server.getDefaultAddressSpaces)
	router.Methods("POST").Path("/IpamDriver.RequestPool").HandlerFunc(server.requestPool)
	router.Methods("POST").Path("/IpamDriver.RequestAddress").HandlerFunc(server.requestAddress)
	router.Methods("POST").Path("/IpamDriver.ReleaseAddress").HandlerFunc(server.releaseAddress)
	router.Methods("POST").Path("/IpamDriver.ReleasePool").HandlerFunc(server.releasePool)
	

    log.Info("Serving Requests")

    return http.Serve(socket, router)
}

type activateResp struct {
	Implements []string
}

func activate(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing Activation Request")
    resp := &activateResp {
		[]string{"NetworkDriver","IpamDriver"},
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		sendError(w, "encode error", http.StatusInternalServerError)
		return
	}
}

// Driver invocations ---

func (server *server) getCapabilities(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing GetCapabilities Request")
	caps, err := server.d.GetCapabilities()
	objectOrErrorResponse(w, caps, err)
}

func (server *server) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create netApi.CreateNetworkRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.CreateNetwork(&create))
}

func (server *server) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete netApi.DeleteNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.DeleteNetwork(&delete))
}

func (server *server) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create netApi.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	res, err := server.d.CreateEndpoint(&create)
	objectOrErrorResponse(w, res, err)
}

func (server *server) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete netApi.DeleteEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.DeleteEndpoint(&delete))
}

func (server *server) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var req netApi.EndpointInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	info, err := server.d.EndpointInfo(&req)
	objectOrErrorResponse(w, info, err)
}

func (server *server) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var join netApi.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&join); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	res, err := server.d.JoinEndpoint(&join)
	objectOrErrorResponse(w, res, err)
}

func (server *server) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l netApi.LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.LeaveEndpoint(&l))
}

func (server *server) getIPAMCapabilities(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing IPAM GetCapabilities Request")
	caps, err := server.d.GetIPAMCapabilities()
	objectOrErrorResponse(w, caps, err)
}

func (server *server) getDefaultAddressSpaces(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing getDefaultAddressSpaces Request")
	
	spaces, err := server.d.GetDefaultAddressSpaces()
	objectOrErrorResponse(w, spaces, err)
}

func (server *server) requestPool(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing requestPool Request")
	var pool ipamApi.RequestPoolRequest
	if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	
	res, err := server.d.RequestPool(&pool)
	objectOrErrorResponse(w, res, err)
}

func (server *server) requestAddress(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing requestAddress Request")
	var address ipamApi.RequestAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	
	res, err := server.d.RequestAddress(&address)
	objectOrErrorResponse(w, res, err)
}
func (server *server) releaseAddress(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing releaseAddress Request")
	var address ipamApi.ReleaseAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.ReleaseAddress(&address))
}

func (server *server) releasePool(w http.ResponseWriter, r *http.Request) {
	log.Info("Processing releasePool Request")
	var pool ipamApi.ReleasePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	emptyOrErrorResponse(w, server.d.ReleasePool(&pool))
}

// Message processing

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Warnf("plugin Not found: [ %+v ]", r)
	http.NotFound(w, r)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	log.Errorf("%d %s", code, msg)
	http.Error(w, msg, code)
}

func errorResponse(w http.ResponseWriter, fmtString string, item ...interface{}) {
	json.NewEncoder(w).Encode(map[string]string{
		"Err": fmt.Sprintf(fmtString, item...),
	})
}

func objectResponse(w http.ResponseWriter, obj interface{}) {
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		sendError(w, "Could not JSON encode response", http.StatusInternalServerError)
		return
	}
}

func emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(map[string]string{})
}

func objectOrErrorResponse(w http.ResponseWriter, obj interface{}, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	objectResponse(w, obj)
}

func emptyOrErrorResponse(w http.ResponseWriter, err error) {
	if err != nil {
		errorResponse(w, err.Error())
		return
	}
	emptyResponse(w)
}