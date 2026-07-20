#!/usr/bin/env bash

kubectl delete deployment auth
kubectl delete deployment api-gateway
kubectl delete deployment analytics
kubectl delete deployment cart
kubectl delete deployment catalog
kubectl delete deployment inventory
kubectl delete deployment notification
kubectl delete deployment order
kubectl delete deployment payment
kubectl delete deployment pricing
kubectl delete deployment recommendation
kubectl delete deployment review
kubectl delete deployment search
kubectl delete deployment shipping
kubectl delete deployment user
kubectl delete deployment wishlist

# bash /home/bodya/Projects/embedded-market-infrastructure/scripts/delete-deployments.sh