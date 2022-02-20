# Wireguard operator

WIP

# Features 
* Uses userspace implementation of wireguard through [wireguard-go](https://github.com/WireGuard/wireguard-go) 
* Does not need persistance. peer/server keys are stored as k8s secrets and loaded into the wireguard pod
* Exposes a metrics endpoint by utilizing [prometheus_wireguard_exporter](https://github.com/MindFlavor/prometheus_wireguard_exporter)




# installation: 
`
operator-sdk run bundle ghcr.io/jodevsa/wireguard-operator-operator-bundle:main
`
