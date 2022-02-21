# Wireguard operator
painless deployment of wireguard on kubernetes

# Support and discussions


Currently the opeartor has only been tested on GKE; if you are facing any problems please open an issue or join our [slack channel](https://join.slack.com/t/wireguard-operator/shared_invite/zt-144xd8ufl-NvH_T82QA0lrP3q0ECTdYA)


# Architecture 

![alt text](./readme/main.png)
# Features 
* Uses userspace implementation of wireguard through [wireguard-go](https://github.com/WireGuard/wireguard-go) 
* Automatic key generation
* Automatic IP allocation
* Does not need persistance. peer/server keys are stored as k8s secrets and loaded into the wireguard pod
* Exposes a metrics endpoint by utilizing [prometheus_wireguard_exporter](https://github.com/MindFlavor/prometheus_wireguard_exporter)

# Example

## server 
```
apiVersion: vpn.example.com/v1alpha1
kind: Wireguard
metadata:
  name: "my-cool-vpn"
spec:
  mtu: "1380"
```


## peer

```
apiVersion: vpn.example.com/v1alpha1
kind: WireguardPeer
metadata:
  name: peer1
spec:
  wireguardRef: "my-cool-vpn"

```



### Peer configuration

Peer configuration can be retreived using the following command
#### command:
```
kubectl get wireguardpeer peer1 --template={{.status.config}} | bash
```
#### output:
```
[Interface]
PrivateKey = WOhR7uTMAqmZamc1umzfwm8o4ZxLdR5LjDcUYaW/PH8=
Address = 10.8.0.3
DNS = 1.1.1.1
MTU = 1380

[Peer]
PublicKey = sO3ZWhnIT8owcdsfwiMRu2D8LzKmae2gUAxAmhx5GTg=
AllowedIPs = 0.0.0.0/0
Endpoint = 32.121.45.102:51820
```


# installation: 
`
git clone https://github.com/jodevsa/wireguard-operator
make deploy
`
