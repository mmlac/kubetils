## Running the admission controller

- It is assumed that the pod is running inside the cluster it is controlling
- 
- Create a network policy only allowing ingress from the master subnet
  [https://stackoverflow.com/a/56494510] (if you have NetworkPolicies enabled)
- Sign the TLS cert with the Kubernetes CA 
  [https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74#e859](Diving
  into Kubernetes MutatingAdmissionWebhook)
