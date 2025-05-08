# operator
Operator for Dummy object

## Description

This project implements a custom Kubernetes Operator using the Operator SDK and Go, it has the following features:

* Custom Resource Definition (CRD)

* Controller reconciliation loop

* Status updates

* Ownership and garbage collection

It serves as a hands-on exercise for learning and demonstrating Kubernetes operator development.

### Controller Logic

The DummyReconciler implements the following steps:

- Reconciliation Trigger: Watches for changes to Dummy resources.

Pod Management:

- Creates an nginx Pod for each Dummy if it does not exist.

- Associates the Pod using SetControllerReference.

Status Update:

- Copies spec.message to status.specEcho

- Retrieves and writes the current status.phase of the Pod to status.podStatus

### Custom Resource Definition

```
apiVersion: interview.com/v1alpha1
kind: Dummy
metadata:
  name: dummy1
  namespace: default
spec:
  message: "I'm just a dummy"
```

Fields:

1. spec.message: A string to store the custom message

2. status.specEcho: Mirrors the spec.message

3. status.podStatus: Reflects the lifecycle phase of the associated Pod (e.g., Pending, Running)

## Getting Started

### To Deploy on the cluster
```
    minikube start
    make kustomize 
    make deploy
    kubectl apply -f config/samples/v1alpha1_dummy.yaml
```
### To Uninstall
**Delete the instances (CRs) from the cluster:**

```
    kubectl delete -f config/samples/v1alpha1_dummy.yaml
    make undeploy
```

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

