## Running the admission controller

- It is assumed that the pod is running inside the cluster it is controlling
- 
- Create a network policy only allowing ingress from the master subnet
  [https://stackoverflow.com/a/56494510] (if you have NetworkPolicies enabled)
- Sign the TLS cert with the Kubernetes CA 
  [https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74#e859](Diving
  into Kubernetes MutatingAdmissionWebhook)
  
## Configuration File
General configuration file containing all settings necessary for the application
to run  
Location: `/etc/ipsa/config.yaml`  

Config Format in YAML:
```
application: #reserved but unused for now
imagePullSecretRules:
    "namespaceRegex":
        "imageRegex": ["list of secrets to add to imagePullSecrets array in PodSpec"]
    "default":
        "us.gcr.io/.*":
        - "gcr-secret"]
        "eu.gcr.io/.*":
        - "gcr-eu-secret"
    ".*":
        ".*":
        - "dockerhub-default-credentials"
```
