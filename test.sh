#!/bin/bash

eval $(minikube -p minikube docker-env)
docker build -t minikube/local-test .

kubectl create namespace secret-namespace
kubectl apply -f test/rbac.yaml -n secret-namespace
kubectl apply -f test/pod.yaml -n secret-namespace

i=0
while [ $i -lt 30 ] # Attempt a max of 30 times
do
    secret_exists=$(kubectl get secret mostest_secret -n secret-namespace)
    if [[ $? == 0 ]]; then
        exit 0
    fi

    sleep 1
    ((i++))
done

echo "Secret was not created. The following secrets exist: "
kubectl get secret -n secret-namespace
exit 1
