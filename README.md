

# Wireguard operator
<img width="1394" alt="Screenshot 2022-02-26 at 02 05 29" src="https://user-images.githubusercontent.com/14154314/177223431-445fbbb1-ff5b-4fd5-86b3-850b81f0a98f.png">

painless deployment of wireguard on kubernetes

# Support and discussions

If you are facing any problems please open an issue or join our [slack channel](https://join.slack.com/t/wireguard-operator/shared_invite/zt-144xd8ufl-NvH_T82QA0lrP3q0ECTdYA)

# Tested with
- [x] IBM Cloud Kubernetes Service
- [x] Gcore Labs KMP
requires enable IP forwarding manually. will dig more into automating it but I was able to make it work manually.
- [x] Google Kubernetes Engine
  * requires `spec.mtu: "1380"`
- [x] DigitalOcean Kubernetes
  * requires `spec.serviceType: "NodePort"`. DigitalOcean LoadBalancer does not support UDP. 
- [x] Self managed k3s on Hetzner 
- [ ] Amazon EKS
- [ ] Azure Kubernetes Service
- [ ] ...?

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
DNS = 10.48.0.10, default.svc.cluster.local
MTU = 1380

[Peer]
PublicKey = sO3ZWhnIT8owcdsfwiMRu2D8LzKmae2gUAxAmhx5GTg=
AllowedIPs = 0.0.0.0/0
Endpoint = 32.121.45.102:51820
```


# installation: 
```
kubectl apply -f https://raw.githubusercontent.com/jodevsa/wireguard-operator/0.0.3/release.yaml
```



# uninstall
```
kubectl delete -f https://raw.githubusercontent.com/jodevsa/wireguard-operator/0.0.3/release.yaml
```
