package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"errors"
	"strings"
	"encoding/json"
	"k8s.io/api/admission/v1beta1"

)



var defaultConfig Config = Config {
	imagePullSecretRules: map[string]map[string]string {
		".*": map[string]string {".*": "testSecret"},
	},
}

func kubeSystemDefaultBody() string {
	review := &v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			UID: "test-uid",
		},
	}
	js, err := json.Marshal(review)
	if err != nil {
		panic("Unable to Marshal the kubeSystemDefaultBody AdmissionReview object to JSON")
	}
	return string(js)
}


// io.Reader that returns an error to test the body not being
// able to be read
type errReader int
func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}



// Test the 'happy path' of the HTTP handling code without testing the
// functionality of the admission handler
func blankFuncMux(config Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(config,
		func(*v1beta1.AdmissionRequest, Config) ([]patchOperation, error){
			return nil, nil}))
	return mux
}


// Create a request handler for the Mux we use in the server and apply the request
// to it. Return the response recorder for evaluation of the result.
func makeRequest(request *http.Request, conf Config) *httptest.ResponseRecorder  {
	recorder := httptest.NewRecorder()
	handler := Mux(conf)
	handler.ServeHTTP(recorder, request)
	return recorder
}

// Ensure all requests with verbs other than POST are rejected
func TestForbiddenHttpVerbs(t *testing.T) {
	const wantStatus, wantString = http.StatusMethodNotAllowed, "invalid method"
	config := defaultConfig

	notAllowedVerbs := []string {"GET", "HEAD", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"}

	for _, verb := range notAllowedVerbs {
		t.Run(verb, func(t *testing.T) {
			req, err := http.NewRequest(verb, "/mutate", nil)
			if err != nil {
				t.Fatalf("Error creating the request: %v", err)
			}

			recorder := makeRequest(req, config)

			if status := recorder.Code; status != wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, wantStatus)
			}
			if body := recorder.Body.String(); !strings.Contains(body, wantString) {
				t.Errorf("handler returned wrong body: got '%v' want containing '%v'",
					body, wantString)
			}
		})
	}
}



// Tests a proper return if the server fails to read the body
func TestEmptyBody(t *testing.T) {
	const wantStatus, wantString = http.StatusBadRequest, "could not read request body"
	config := defaultConfig

	req, err := http.NewRequest("POST", "/mutate", errReader(0))
	if err != nil {
		t.Fatalf("Error creating the request: %v", err)
	}

	recorder := makeRequest(req, config)

	if status := recorder.Code; status != wantStatus {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, wantStatus)
	}
	if body := recorder.Body.String(); !strings.Contains(body, wantString) {
		t.Errorf("handler returned wrong body: got '%v' want containing '%v'",
			body, wantString)
	}
}



// Tests that the wrong Content Type will be rejected
func TestWrongContentType(t *testing.T) {
	const wantStatus, wantString = http.StatusBadRequest, "unsupported content type"
	config := defaultConfig

	headers := map[string]map[string][]string{
		"Empty Header:":    map[string][]string{},
		"XML Content Type": map[string][]string{"Content-Type": {"application/xml"}},
	}

	for name, hdr := range headers {
		t.Run(name, func(t *testing.T){
			req, err := http.NewRequest("POST", "/mutate", strings.NewReader(""))
			if err != nil {
				t.Fatalf("Error creating the request: %v", err)
			}

			req.Header = hdr

			recorder := makeRequest(req, config)

			if status := recorder.Code; status != wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, wantStatus)
			}
			if body := recorder.Body.String(); !strings.Contains(body, wantString) {
				t.Errorf("handler returned wrong body: got '%v' want containing '%v'",
					body, wantString)
			}

		})
	}
}

// Tests a properly formed request
func TestProperlyFormedRequest(t *testing.T) {
	const wantStatus, wantString = http.StatusOK, "\"allowed\":true"
	config := defaultConfig

	req, err := http.NewRequest("POST", "/mutate", strings.NewReader(kubeSystemDefaultBody()))
	if err != nil {
		t.Fatalf("Error creating the request: %v", err)
	}

	req.Header = map[string][]string{"Content-Type": {"application/json"}}

	recorder := makeRequest(req, config)

	if status := recorder.Code; status != wantStatus {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, wantStatus)
	}
	if body := recorder.Body.String(); !strings.Contains(body, wantString) {
		t.Errorf("handler returned wrong body: got '%v' want containing '%v'",
			body, wantString)
	}
}




// func TestValidKubeSystemRequest(t *testing.T) {
// 	config  := defaultConfig
// 	content := kubeSystemRequest

// 	req, err := http.NewRequest("GET", "/health-check", nil)
// 	if err != nil {
// 		t.Fatalf("Error creating the request: %v", err)
// 	}


// 	recorder := makeRequest(req, config)

// 	// Check the status code is what we expect.
// 	if status := recorder.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v",
// 			status, http.StatusOK)
// 	}


// 	// Check the response body is what we expect.
// 	expected := `{"alive": true}`
// 	if recorder.Body.String() != expected {
// 		t.Errorf("handler returned unexpected body: got %v want %v",
// 			recorder.Body.String(), expected)
// 	}

// }
