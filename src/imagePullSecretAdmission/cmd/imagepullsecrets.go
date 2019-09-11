/*
Copyright (c) 2019 Markus Lachinger. All rights reserved.
Licensed under the MIT license. See LICENSE file in the project root for details.
*/

package main


import (
	"fmt"
	"regexp"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
)

// Remove user-provided image pull secrets and add managed ones based on configuration.
// This allows also blocking certain registries / paths from specific namespaces.
//
// Examples of use-cases can be found in the tests:  TODO O:)
func manageImagePullSecrets(req *v1beta1.AdmissionRequest, config Config) ([]patchOperation, error) {
	// This handler should only get called on Pod objects as per the MutatingWebhookConfiguration in the YAML file.
	// However, if (for whatever reason) this gets invoked on an object of a different kind, issue a log message but
	// let the object request pass through otherwise.
	if req.Resource != podResource {
		log.Printf("expect resource to be %s", podResource)
		return nil, nil
	}


	// Parse the Pod object.
	raw       := req.Object.Raw
	pod       := corev1.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	var patches []patchOperation
	namespace := req.Namespace

	// Ignore system namespaces
	if namespace == metav1.NamespacePublic || namespace == metav1.NamespaceSystem  || namespace == "istio-system" {
		return nil, nil
	}

	images    := getUniquePodImages(pod)
	patches    = append(patches, removeExistingPullSecrets(namespace, pod)...)

	if config.imagePullSecretRules != nil {
		patches = append(patches, patchPod(config.imagePullSecretRules, namespace, images)...)
	}

	return patches, nil
}


// Remove any ImagePullSecret that the user has added.
// The idea is that only managed image pull secrets are allowed.
func removeExistingPullSecrets(ns string, pod corev1.Pod) []patchOperation {
	if len(pod.Spec.ImagePullSecrets) == 0 {
		return nil
	} else {
		return []patchOperation{patchOperation{Op: "remove", Path: "/spec/ImagePullSecrets"}}
	}
}


// Iterates through all containers and initContainers of the Pod
// and outputs a unique list of images this pod uses
func getUniquePodImages(pod corev1.Pod) []string {
	// Use a map key-assignment as a uniqueness-check for images
	var imageMap map[string]struct{}

	//return as slice of unique images
	var imageSlice []string


	for _, container := range pod.Spec.Containers {
		imageMap[container.Image] = struct{}{}
	}

	for _, container := range pod.Spec.InitContainers {
		imageMap[container.Image] = struct{}{}
	}



	for image, _ := range imageMap {
		imageSlice = append(imageSlice, image)
	}

	return imageSlice
}


// Takes all images this pod uses and matches it against the rules in the config.
// Adds a unique set of imagePullSecrets as directed by the rules.
func patchPod(imagePullSecretRules map[string]map[string]string, namespace string, images []string) []patchOperation {
	var secretsMap map[string]struct{}
	var patches []patchOperation

	// We need to create a fresh ImagePullSecrets array
	// because we removed it with a patch beforehand or
	// expect it to not exist
	patches = append(patches, patchOperation {
		Op: "add",
			Path: "/spec/ImagePullSecrets",
			Value: "[]",
		})


	for namespaceRegex, imageMap := range imagePullSecretRules {
		match, _ := regexp.MatchString(namespaceRegex, namespace)
		if match {
			for imageRegex, imagePullSecret := range imageMap {
				for _, currentImage := range images {
					imageMatch, _ := regexp.MatchString(imageRegex, currentImage)
					if imageMatch {
						secretsMap[imagePullSecret] = struct{}{}
					}
				}
			}
		}
	}

	for secret, _ := range secretsMap {
		patches = append(patches, patchOperation{
			Op: "add",
			Path: "/spec/ImagePullSecrets/-",
			Value: secret,
		})
	}


	return patches
}

