Add repo: helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
Update: helm repo update

install: helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace -f helm-values.yaml

upgrade: helm upgrade ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx -f helm-values.yaml --dry-run --debug