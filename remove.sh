kubectl delete deployment wireguard-operator-controller-manager
kubectl delete catalogsource.operators.coreos.com wireguard-operator-catalog
kubectl delete subscriptions.operators.coreos.com wireguard-operator-v0-0-1-sub
kubectl delete csv wireguard-operator.v0.0.1