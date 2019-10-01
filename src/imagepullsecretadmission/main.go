/*
Copyright (c) 2019 Markus Lachinger. All rights reserved.
Licensed under the MIT license. See LICENSE file in the project root for details.
*/

/*  TODOS / FUTURE POSSIBLE FEATURES

    TODO tests O:)
    TODO Add ability to exclude more namespaces via config
    TODO Add ability to have an override flag for removing pull secrets. Needs another admission
         controller to manage who is allowed to add these annotations or use it as emergency flag
         under discretion.
    TODO Have a second set of rules that does image admission itself, i.e. has a regex of
         namespaceRegex -> imageRegex
         if there is no match, reject the pod.
         If YAML is guaranteed to keep the order in the parsing we can do more complex allow/deny
         override rules. Probably not "most specific wins"
         To reject a request, our controller logic needs to return an error back to the handling logic
         import "errors"   errors.New("error message")
    TODO Consider wrapping all dependencies into a server type
*/

package main

import (
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
	Application          map[string]string  `yaml:"application,omitempty"`
	ImagePullSecretRules map[string]map[string]string `yaml:"imagePullSecretRules"`
}



func Mux(config Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(config, manageImagePullSecrets))
	return mux
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

	mux := Mux(config)
 	server := &http.Server{
		// We listen on port 8443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":8443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}
