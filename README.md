# Wireguard operator

![alt text](./readme/main.png)
# Features 
* Uses userspace implementation of wireguard through [wireguard-go](https://github.com/WireGuard/wireguard-go) 
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
  name: peernew
spec:
  wireguardRef: "my-cool-vpn"

```



# installation: 
`
operator-sdk run bundle ghcr.io/jodevsa/wireguard-operator-operator-bundle:main
`
