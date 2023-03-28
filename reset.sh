kubectl apply -f k8s/01-base/color-mapping.yaml
kubectl apply -f k8s/01-base/color-profile.yaml
kubectl apply -f k8s/01-base/smiley-mapping.yaml
kubectl apply -f k8s/01-base/smiley-profile.yaml

kubectl apply -f k8s/01-base/face-mapping.yaml
kubectl apply -f k8s/01-base/face-profile.yaml

kubectl apply -f k8s/01-base/faces-gui-mapping.yaml

kubectl delete -n faces deploy/face || true
linkerd inject k8s/01-base/faces.yaml | kubectl apply --overwrite -f -
