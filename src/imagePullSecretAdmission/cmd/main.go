/*
Copyright (c) 2019 Markus Lachinger. All rights reserved.
Licensed under the MIT license. See LICENSE file in the project root for details.
*/

/*  TODOS / FUTURE POSSIBLE FEATURES

    TODO tests O:)
    TODO Add ability to exclude more custom spaces in #removeExistingPullSecrets via config file
    TODO Add ability to have an override flag for removing pull secrets. Needs another admission
         controller to manage who is allowed to add these annotations or use it as emergency flag
         under discretion.
*/

package main

import (
	"errors"
	"fmt"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"path/filepath"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)


const (
	tlsDir      = `/run/secrets/tls`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
	configFile  = `/etc/ipsa/config.yaml`
)

var (
	podResource = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
)

type Config struct {
	namespaceRestrictions map[string]string
}







// Remove user-provided image pull secrets and add managed ones based on configuration.
// This allows also blocking certain registries / paths from specific namespaces.
//
// Examples of use-cases can be found in the tests:  TODO O:)
func manageImagePullSecrets(req *v1beta1.AdmissionRequest) ([]patchOperation, error) {
	// Slice of patches being returned
	var patches []patchOperation

	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}


	// Parse the Pod object.
	raw       := req.Object.Raw
	namespace := req.Namespace
	pod       := corev1.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	patches = append(patches, removeExistingPullSecrets(namespace, pod)...)


	// filter by config entries

	// create patches for pod based on rules

	return patches, nil
}

// Remove any ImagePullSecret that the user has added, unless it's in the Kubernetes or istio system spaces
// The idea is that only managed image pull secrets are allowed.
func removeExistingPullSecrets(ns string, pod corev1.Pod) []patchOperation {
	if ns == metav1.NamespacePublic || ns == metav1.NamespaceSystem  || ns == "istio-system" {
		return nil
	} else if len(pod.Spec.ImagePullSecrets) == 0 {
		return nil
	} else {
		return []patchOperation{patchOperation{Op: "remove", Path: "/spec/ImagePullSecrets"}}
	}
}



// applySecurityDefaults implements the logic of our example admission controller webhook. For every pod that is created
// (outside of Kubernetes namespaces), it first checks if `runAsNonRoot` is set. If it is not, it is set to a default
// value of `false`. Furthermore, if `runAsUser` is not set (and `runAsNonRoot` was not initially set), it defaults
// `runAsUser` to a value of 1234.
//
// To demonstrate how requests can be rejected, this webhook further validates that the `runAsNonRoot` setting does
// not conflict with the `runAsUser` setting - i.e., if the former is set to `true`, the latter must not be `0`.
// Note that we combine both the setting of defaults and the check for potential conflicts in one webhook; ideally,
// the latter would be performed in a validating webhook admission controller.
func applySecurityDefaults(req *v1beta1.AdmissionRequest) ([]patchOperation, error) {
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}

	// What namespaces are allowed to load images from where
	// TODO: load this from a YAML file
	// var namespaceRestrictions = map[string]string{"*": "*"}

	// Implement a filter. Input is namespace-regex -> image url regex
	// remove any imagePullSecrets existing
	// set imagepullsecrets based on image domain(s) (can be multiple containers)
		//var namespace string = req.Namespace


	// Parse the Pod object.
	raw := req.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	// Get imagePullSecrets
	// ips := pod.Spec.ImagePullSecrets

	// Retrieve the `runAsNonRoot` and `runAsUser` values.
	var runAsNonRoot *bool
	var runAsUser *int64
	if pod.Spec.SecurityContext != nil {
		runAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
		runAsUser = pod.Spec.SecurityContext.RunAsUser
	}

	// Create patch operations to apply sensible defaults, if those options are not set explicitly.
	var patches []patchOperation
	if runAsNonRoot == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/securityContext/runAsNonRoot",
			// The value must not be true if runAsUser is set to 0, as otherwise we would create a conflicting
			// configuration ourselves.
			Value: runAsUser == nil || *runAsUser != 0,
		})

		if runAsUser == nil {
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/spec/securityContext/runAsUser",
				Value: 1234,
			})
		}
	} else if *runAsNonRoot == true && (runAsUser != nil && *runAsUser == 0) {
		// Make sure that the settings are not contradictory, and fail the object creation if they are.
		return nil, errors.New("runAsNonRoot specified, but runAsUser set to 0 (the root user)")
	}

	return patches, nil
}

// Start http server, pass request through admissionFuncHandler to parse request,
// run applySecurityDefaults function and form the proper HTTP response.
func main() {
	configFileContent, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Cannot read config file from file %s: %s. Aborting...", configFile, err.Error())
	}

	var config Config
	yaml.Unmarshal(configFileContent, &config)

	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath  := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(config, applySecurityDefaults))
	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":8443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}
