# Wireguard Operator
<img width="1394" alt="Screenshot 2022-02-26 at 02 05 29" src="https://user-images.githubusercontent.com/14154314/177223431-445fbbb1-ff5b-4fd5-86b3-850b81f0a98f.png">

Painless deployment of wireguard on kubernetes

## Support and discussions

If you are facing any problems please open an [issue](https://github.com/jodevsa/wireguard-operator/issues) or start a 
[discussion](https://github.com/jodevsa/wireguard-operator/discussions) 

## Tested with
- [x] IBM Cloud Kubernetes Service
- [x] Gcore Labs KMP
  * requires `spec.enableIpForwardOnPodInit: true`
- [x] Google Kubernetes Engine
  * requires `spec.mtu: "1380"`
  * Not compatible with "Container-Optimized OS with containerd" node images
  * Not compatible with autopilot
- [x] DigitalOcean Kubernetes
  * requires `spec.serviceType: "NodePort"`. DigitalOcean LoadBalancer does not support UDP.
- [ ] Amazon EKS
- [ ] Azure Kubernetes Service
- [x] Hetzner Cloud
  * requires `spec.serviceType: "NodePort"`. Hetzner Cloud LoadBalancer does not support UDP.
- [x] Vultr
- [x] Proxmox
- [ ] ...?

## Architecture 

![alt text](./readme/main.png)

## Features 
* Falls back to userspace implementation of wireguard [wireguard-go](https://github.com/WireGuard/wireguard-go) if wireguard kernal module is missing
* Automatic key generation
* Automatic IP allocation
* Does not need persistance. peer/server keys are stored as k8s secrets and loaded into the wireguard pod
* Exposes a metrics endpoint by utilizing [prometheus_wireguard_exporter](https://github.com/MindFlavor/prometheus_wireguard_exporter)

## Example

### Server

```
apiVersion: vpn.wireguard-operator.io/v1alpha1
kind: Wireguard
metadata:
  name: "my-cool-vpn"
spec:
  mtu: "1380"
```

### Peer

```
apiVersion: vpn.wireguard-operator.io/v1alpha1
kind: WireguardPeer
metadata:
  name: peer1
spec:
  wireguardRef: "my-cool-vpn"
```

#### Peer configuration

Peer configuration can be retrieved using the following command:

```console
kubectl get wireguardpeer peer1 --template={{.status.config}} | bash
```

After executing it, something similar to the following will be shown. Use this config snippet to configure your
preferred Wireguard client:

```console
[Interface]
PrivateKey = WOhR7uTMAqmZamc1umzfwm8o4ZxLdR5LjDcUYaW/PH8=
Address = 10.8.0.3
DNS = 10.48.0.10, default.svc.cluster.local
MTU = 1380

[Peer]
PublicKey = sO3ZWhnIT8owcdsfwiMRu2D8LzKmae2gUAxAmhx5GTg=
AllowedIPs = 0.0.0.0/0
Endpoint = 32.121.45.102:51820
```

## How to deploy
```
kubectl apply -f https://github.com/jodevsa/wireguard-operator/releases/download/v2.7.0/release.yaml
```

## How to remove
```
kubectl delete -f https://github.com/jodevsa/wireguard-operator/releases/download/v2.7.0/release.yaml
```

## How to collaborate

This project is done on top of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), so read about that project
before collaborating. Of course, we are open to external collaborations for this project. For doing it you must fork the
repository, make your changes to the code and open a PR. The code will be reviewed and tested (always)

> We are developers and hate bad code. For that reason we ask you the highest quality on each line of code to improve
> this project on each iteration.
