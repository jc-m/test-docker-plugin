#### Experimental (LOCAL) docker libnetwork plugin ####

This is not a funtional plugin but mostly the implementation of the libnetwork NetworkDriver and IpamDriver interfaces simulating the behavior sufficiently to please the docker daemon.

The driver is created and uploaded in a docker container run on the docker-machine along with containers it configures.

Note that this driver works only with docker built with this branch :
https://github.com/jc-m/docker/tree/IPAliases
and
https://github.com/jc-m/libnetwork/tree/IPAliases

to run the container for testing :
```
docker run -ti --privileged --net=host --rm -v /run/docker/plugins:/run/docker/plugins jc-m/routed-driver -log-level debug

```

run in another shell the commands like :

```
docker network create --internal --driver=routed --ipam-driver=routed --subnet 10.46.0.0/16  mine

docker run -ti --net=mine --ip 10.46.1.7 --label=com.docker.network.ip_aliases=10.255.255.254/32 --rm alpine-test sh

docker network rm mine

```
