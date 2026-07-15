h install: helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
h update: helm repo update
h install-sealed-secrets: helm install
	helm upgrade --install sealed-secrets sealed-secrets/sealed-secrets \
		--namespace kube-system \
		--set rbac.create=true \
		--set serviceAccount.create=true \
		--set serviceAccount.name=sealed-secrets-controller \
		--set image.repository=quay.io/bitnami/sealed-secrets-controller \
		--set image.tag=v0.27.3

create secret:  kubectl \
	create secret generic argocd-repo-creds --from-file=sshPrivateKey=/home/bodya/.ssh/github/market-infrastructure/id_ed25519
    --dry-run=client -o yaml | kubeseal --cert my-pub-key.pem >  \
	gitops-manifests/secrets/argocd-repo-secret.yaml 