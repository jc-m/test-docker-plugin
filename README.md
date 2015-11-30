#### Experimental (LOCAL) docker libnetwork plugin ####

This is not a funtional plugin but mostly the implementation of the libnetwork NetworkDriver and IpamDriver interfaces simulating the behavior sufficiently to please the docker daemon.

The driver is created and uploaded in a docker container run on the docker-machine along with containers it configures.

Note that this driver needs the following patch to libnetwork/default_gateway.go :
```
func (sb *sandbox) needDefaultGW() bool {
	var needGW bool

	for _, ep := range sb.getConnectedEndpoints() {
		if ep.endpointInGWNetwork() {
			continue
		}
		if ep.getNetwork().Type() != "bridge" &&
		ep.getNetwork().Type() != "overlay" &&
		ep.getNetwork().Type() != "windows" {
				continue
		}
		// TODO v6 needs to be handled.
		if len(ep.Gateway()) > 0 {
			return false
		}
		needGW = true
	}
	return needGW
}
```

to run the container for testing :
```
docker run -ti --privileged --net=host --rm -v /run/docker/plugins:/run/docker/plugins jc-m/routed-driver
```

run in another shell the commands like :
```
docker network create --driver=routed --ipam-driver=routed  mine


INFO[0000] Test routed network plugin
INFO[0000] Serving Requests
INFO[0005] Processing requestPool Request
INFO[0005] Pool Request request: &{AddressSpace:Testlocal Pool: SubPool: Options:map[] V6:false}
INFO[0005] Pool Request: responded with &{Response:{Error:} PoolID:myPool Pool:100.64.0.0/10 Data:map[]}
INFO[0005] Processing requestAddress Request
INFO[0005] Address Request request: &{PoolID:myPool Address: Options:map[]}
INFO[0005] addresse request: responded with &{Response:{Error:} Address:100.64.0.134/32 Data:map[]}
INFO[0005] Create network request &{NetworkID:7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23 Options:map[com.docker.network.generic:map[]] IPv4Data:[{AddressSpace: Pool:100.64.0.0/10 Gateway:100.64.0.134/32 AuxAddresses:map[]}] IPv6Data:[]}
INFO[0005] Create network 7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23


docker run -ti --net=mine --rm alpine-test sh

INFO[0026] Processing requestAddress Request
INFO[0026] Address Request request: &{PoolID:myPool Address: Options:map[]}
INFO[0026] addresse request: responded with &{Response:{Error:} Address:100.64.0.238/32 Data:map[]}
INFO[0026] Create endpoint request &{NetworkID:7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23 EndpointID:5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2 Interface:0x186b6c40 Options:map[com.docker.network.endpoint.exposedports:[] com.docker.network.portmap:[]]}
INFO[0026] Requested Interface &{Address:100.64.0.238/32 AddressIPv6: MacAddress:}
INFO[0026] Create endpoint 5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2 &{Response:{Err:} Interface:0x186b6dc0}
INFO[0026] Delete endpoint request: 0x1860ab30
INFO[0026] Join endpoint 7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23:5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2 to /var/run/docker/netns/1870918c9aec

After terminating the container :

INFO[0033] Leave request: 0x1860ac60
INFO[0033] Leave 7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23:5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2
INFO[0033] Delete endpoint request: &{NetworkID:7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23 EndpointID:5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2}
INFO[0033] Delete endpoint 5870ffa610cf8d224d9e32f1dbb0d11b609bd83ebd9c20e488ac57accfa479c2
INFO[0033] Processing releaseAddress Request
INFO[0033] Address Release request: &{PoolID:myPool Address:100.64.0.238}
INFO[0033] addresse release 100.64.0.238 from myPool

docker network rm mine

INFO[0044] Delete network request: &{NetworkID:7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23}
INFO[0044] Destroy network 7a5f96b67ae2ebdb681dd3d13b5404501f8f7e87cbe4bba3db770ee5015b8e23
INFO[0044] Processing releaseAddress Request
INFO[0044] Address Release request: &{PoolID:myPool Address:100.64.0.134}
INFO[0044] addresse release 100.64.0.134 from myPool
INFO[0044] Processing releasePool Request
INFO[0044] Pool Release request: &{PoolID:myPool}
INFO[0044] pool release myPool

```