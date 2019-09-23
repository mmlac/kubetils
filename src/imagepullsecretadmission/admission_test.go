package main

import (
	"testing"
	"encoding/json"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)



func TestKubeSystemAdmission(t *testing.T) {


	namespace := []string{"kube-system", "kube-public", "istio-system"}

	config := defaultConfig

	for _, ns := range namespace {
		ns := ns
		t.Run(ns, func(t *testing.T){
			request := &v1beta1.AdmissionRequest {
				UID:        "test-uid",
				Namespace:  ns,
				Resource:   podResource,
				}

			res, err := manageImagePullSecrets(request, config)
			if err != nil {
				t.Errorf("Error: Wanted nil, got %v", err)
			}

			if res != nil {
				t.Errorf("Result: Wanted nil, got %v", res)
			}
		})
	}
}


func TestKubeNormalAdmission(t *testing.T) {
	namespace := []string{"kube-nothing", "test-system", "testns"}

	config := defaultConfig

	for _, ns := range namespace {
		ns := ns
		t.Run(ns, func(t *testing.T){
			var raw runtime.RawExtension
			jsonbytes, err := json.Marshal(corev1.Pod {
				ObjectMeta: metav1.ObjectMeta {
					Namespace: ns,
				},
					Spec: corev1.PodSpec {
					Containers: []corev1.Container{
						corev1.Container {
							Image: "test",
							},
					},
					},
				})
			if err != nil {
				t.Fatalf("Failed JSON marshal with %v", err)
			}

			raw.UnmarshalJSON(jsonbytes)

			request := &v1beta1.AdmissionRequest {
				UID:        "test-uid",
				Namespace:  ns,
				Resource:   podResource,
				Object:     raw,
			}

			res, err := manageImagePullSecrets(request, config)

			if err != nil {
				t.Errorf("Error: Wanted nil, got %v", err)
			}

			if res == nil || len(res) != 2 {
				t.Errorf("Result: Wanted patch result, got %v", res)
			} else {
				if res[0].Op != "add" || res[0].Path != "/spec/imagePullSecrets" || res[0].Value != "[]" {
					t.Errorf("Result: Expected first patch to add empty imagePullSecrets array, got '%v'", res[0])
				}
				if res[1].Op != "add" || res[1].Path != "/spec/imagePullSecrets/-" || res[1].Value != "testSecret" {
					t.Errorf("Result: Expected first patch to add empty imagePullSecrets array, got '%v'", res[0])
				}
			}
		})
	}
}



func TestKubeNotAPodAdmission(t *testing.T) {
	namespace := []string{"kube-nothing", "test-system", "testns"}

	config := defaultConfig


	for _, ns := range namespace {
		ns := ns
		t.Run(ns, func(t *testing.T){
			request := &v1beta1.AdmissionRequest {
				UID:        "test-uid",
					Namespace:  ns,
					Resource: metav1.GroupVersionResource{Version: "v1", Resource: "services"},
				}

			res, err := manageImagePullSecrets(request, config)
			if err != nil {
				t.Errorf("Error: Wanted nil, got %v", err)
			}

			if res != nil {
				t.Errorf("Result: Wanted nil, got %v", res)
			}
		})
	}
}

