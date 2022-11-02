#!/bin/bash

eval $(minikube -p minikube docker-env)
docker build -t minikube/local-test .

kubectl create namespace secret-namespace
kubectl apply -f test/rbac.yaml -n secret-namespace
kubectl apply -f test/pod.yaml -n secret-namespace

i=0
while [ $i -lt 10 ] # Attempt a max of 10 times
do
    secret_exists=$(kubectl get secret mostest-secret -n secret-namespace)
    if [[ $? == 0 ]]; then
        echo "Secret found, exiting."
        exit 0
    fi

    echo "Retrying."
    sleep 1
    ((i++))
done

echo "Secret was not created."
echo "--- Existing secrets ---"
kubectl get secret -n secret-namespace
echo "--- Pod logs -----------"
kubectl logs local-test -n secret-namespace
exit 1
