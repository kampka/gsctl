package client

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/client/clienterror"
	"github.com/giantswarm/gsctl/testutils"
)

// TestRedactArgs tests redactArgs().
func TestRedactArgs(t *testing.T) {
	argtests := []struct {
		in  string
		out string
	}{
		// these remain unchanged
		{"foo", "foo"},
		{"foo bar", "foo bar"},
		{"foo bar blah", "foo bar blah"},
		{"foo bar blah -p mypass", "foo bar blah -p mypass"},
		{"foo bar blah -p=mypass", "foo bar blah -p=mypass"},
		// these will be altered
		{"foo bar blah --password mypass", "foo bar blah --password REDACTED"},
		{"foo bar blah --password=mypass", "foo bar blah --password=REDACTED"},
		{"foo bar blah --auth-token=some-token", "foo bar blah --auth-token=REDACTED"},
		{"foo bar blah --auth-token some-token", "foo bar blah --auth-token REDACTED"},
		{"foo login blah -p mypass", "foo login blah -p REDACTED"},
		{"foo login blah -p=mypass", "foo login blah -p=REDACTED"},
	}

	for _, tt := range argtests {
		in := strings.Split(tt.in, " ")
		out := strings.Join(redactArgs(in), " ")
		if out != tt.out {
			t.Errorf("want '%q', have '%s'", tt.in, tt.out)
		}
	}
}

// TestNoConnection checks out how the latest client deals with a missing
// server connection.
func TestNoConnection(t *testing.T) {
	// a non-existing endpoint (must use an IP, not a hostname)
	config := &Configuration{
		Endpoint: "http://127.0.0.1:55555",
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password", nil)

	if err == nil {
		t.Error("Expected 'connection refused' error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	_, ok = clientAPIError.OriginalError.(*net.OpError)
	if !ok {
		t.Error("Type assertion to *net.OpError not successful")
	}

	t.Logf("clientAPIError: %#v", clientAPIError)

	if clientAPIError.ErrorMessage == "" {
		t.Error("ErrorMessage was empty, expected helpful message.")
	}
	if clientAPIError.ErrorDetails == "" {
		t.Error("ErrorDetails was empty, expected helpful message.")
	}
}

// TestHostnameUnresolvable checks out how the latest client deals with a
// non-resolvable host name.
func TestHostnameUnresolvable(t *testing.T) { // Our test server.

	// a non-existing host name
	config := &Configuration{
		Endpoint: "http://non.existing.host.name",
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	responseBody, err := gsClient.CreateAuthToken("email", "password", nil)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if responseBody != nil {
		t.Errorf("Expected nil response body, got %#v", responseBody)
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	_, ok = clientAPIError.OriginalError.(*net.DNSError)
	if !ok {
		t.Error("Type assertion to *net.DNSError not successful")
	}

	t.Logf("clientAPIError: %#v", clientAPIError)

	if clientAPIError.ErrorMessage == "" {
		t.Error("ErrorMessage was empty, expected helpful message.")
	}
	if clientAPIError.ErrorDetails == "" {
		t.Error("ErrorDetails was empty, expected helpful message.")
	}
}

// TestTimeout tests if the latest client handles timeouts as expected.
func TestTimeout(t *testing.T) {
	// Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// enforce a timeout longer than the client's
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "Hello")
	}))
	defer ts.Close()

	clientConfig := &Configuration{
		Endpoint: ts.URL,
		Timeout:  500 * time.Millisecond,
	}
	gsClient, err := New(clientConfig)
	if err != nil {
		t.Error(err)
	}
	resp, err := gsClient.CreateAuthToken("email", "password", nil)
	if err == nil {
		t.Error("Expected Timeout error, got nil")
		t.Logf("resp: %#v", resp)
	} else {
		clientAPIError, ok := err.(*clienterror.APIError)
		if !ok {
			t.Error("Type assertion err.(*clienterror.APIError) not successful")
		}
		if !clientAPIError.IsTimeout {
			t.Error("Expected clientAPIError.IsTimeout to be true, got false")
		}
	}
}

// TestUserAgent tests whether our user-agent header appears in requests.
func TestUserAgent(t *testing.T) {
	clientConfig := &Configuration{
		UserAgent: "my own user agent/1.0",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("user-agent")

		if ua != clientConfig.UserAgent {
			t.Errorf("Expected '%s', got '%s'", clientConfig.UserAgent, ua)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "NONE", "message": "none"}`))
	}))
	defer ts.Close()

	clientConfig.Endpoint = ts.URL

	gsClient, err := New(clientConfig)
	if err != nil {
		t.Error(err)
	}

	// just issue a request, don't care about the result.
	gsClient.CreateAuthToken("email", "password", nil)
}

// TestForbidden tests out how the latest client gives access to
// HTTP error details for a 403 error.
func TestForbidden(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`Access forbidden`))
	}))
	defer ts.Close()

	gsClient, err := New(&Configuration{Endpoint: ts.URL})
	if err != nil {
		t.Error(err)
	}

	response, err := gsClient.CreateAuthToken("email", "password", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if response != nil {
		t.Error("Expected nil response")
	}

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	if clientAPIError.HTTPStatusCode != http.StatusForbidden {
		t.Error("Expected HTTP status 403, got", clientAPIError.HTTPStatusCode)
	}
}

// TestUnauthorized tests out how the latest client gives access to
// HTTP error details for a 401 error.
func TestUnauthorized(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "PERMISSION_DENIED", "message": "Not authorized"}`))
	}))
	defer ts.Close()

	gsClient, err := New(&Configuration{Endpoint: ts.URL})
	if err != nil {
		t.Error(err)
	}

	_, err = gsClient.DeleteAuthToken("foo", nil)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	t.Logf("err: %#v", err)

	clientAPIError, ok := err.(*clienterror.APIError)
	if !ok {
		t.Error("Type assertion err.(*clienterror.APIError) not successful")
	}

	if clientAPIError.HTTPStatusCode != http.StatusUnauthorized {
		t.Error("Expected HTTP status 401, got", clientAPIError.HTTPStatusCode)
	}
}

// TestAuxiliaryParams checks whether the client carries through our auxiliary
// parameters.
func TestAuxiliaryParams(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-Request-ID") != "request-id" {
			t.Error("Header X-Request-ID not available")
		}
		if r.Header.Get("X-Giant-Swarm-CmdLine") != "command-line" {
			t.Error("Header X-Giant-Swarm-CmdLine not available")
		}
		if r.Header.Get("X-Giant-Swarm-Activity") != "activity-name" {
			t.Error("Header X-Giant-Swarm-Activity not available")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"foo": "bar"}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	ap := gsClient.DefaultAuxiliaryParams()
	ap.RequestID = "request-id"
	ap.CommandLine = "command-line"
	ap.ActivityName = "activity-name"

	gsClient.CreateAuthToken("foo", "bar", ap)
}

// TestCreateAuthToken checks out how creating an auth token works in
// our new client.
func TestCreateAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"auth_token": "e5239484-2299-41df-b901-d0568db7e3f9"}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	response, err := gsClient.CreateAuthToken("foo", "bar", nil)
	if err != nil {
		t.Error(err)
	}

	if response.Payload.AuthToken != "e5239484-2299-41df-b901-d0568db7e3f9" {
		t.Errorf("Didn't get the expected token. Got %s", response.Payload.AuthToken)
	}
}

// TestDeleteAuthToken checks out how to issue an authenticted request
// using the new client.
func TestDeleteAuthToken(t *testing.T) { // Our test server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "giantswarm test-token" {
			t.Error("Bad authorization header:", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": "RESOURCE_DELETED", "message": "The authentication token has been successfully deleted."}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	response, err := gsClient.DeleteAuthToken("test-token", nil)
	if err != nil {
		t.Error(err)
	}

	if response.Payload.Code != "RESOURCE_DELETED" {
		t.Errorf("Didn't get the RESOURCE_DELETED message. Got '%s'", response.Payload.Code)
	}
}

func TestGetClusterStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cluster": {
				"conditions": [
					{
						"status": "True",
						"type": "Created"
					}
				],
				"network": {
					"cidr": ""
				},
				"nodes": [
					{
						"name": "4jr2w-master-000000",
						"version": "2.0.1"
					},
					{
						"name": "4jr2w-worker-000001",
						"version": "2.0.1"
					}
				],
				"resources": [],
				"versions": [
					{
						"date": "0001-01-01T00:00:00Z",
						"semver": "2.0.1"
					}
				]
			}
		}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	status, err := gsClient.GetClusterStatus("cluster-id", nil)
	if err != nil {
		t.Error(err)
	}

	if len(status.Cluster.Nodes) != 2 {
		t.Errorf("Expected status.Nodes to have length 2, but has %d. status: %#v", len(status.Cluster.Nodes), status)
	}
}

func TestGetClusterStatusEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cluster": {
				"nodes": []
			}
		}`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	status, err := gsClient.GetClusterStatus("cluster-id", nil)
	if err != nil {
		t.Error(err)
	}

	if len(status.Cluster.Nodes) != 0 {
		t.Errorf("Expected status.Nodes to have length 0. Has length %d", len(status.Cluster.Nodes))
	}
}

// Test_GetDefaultCluster tests the GetDefaultCluster function
// for the case that only one cluster exists
func Test_GetDefaultCluster(t *testing.T) {
	// returns one cluster
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
      {
        "create_date": "2017-04-16T09:30:31.192170835Z",
        "id": "cluster-id",
        "name": "Some random test cluster",
				"owner": "acme"
      }
    ]`))
	}))
	defer mockServer.Close()

	// config
	yamlText := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL
	fs := afero.NewMemMapFs()
	_, err := testutils.TempConfig(fs, yamlText)
	if err != nil {
		t.Error(err)
	}

	clientWrapper, err := NewWithConfig(mockServer.URL, "")

	clusterID, err := clientWrapper.GetDefaultCluster(nil)
	if err != nil {
		t.Error(err)
	}
	if clusterID != "cluster-id" {
		t.Errorf("Expected 'cluster-id', got %#v", clusterID)
	}
}

func TestMalformedResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>This is not a JSON page</html>`))
	}))
	defer mockServer.Close()

	config := &Configuration{
		Endpoint: mockServer.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	_, err = gsClient.GetClusterV4("cluster-id", nil)
	if !clienterror.IsMalformedResponse(err) {
		t.Errorf("Expected 'Malformed response' error, got %s", err.Error())
	}
}

// Test_CertificateSignedByUnknownAuthority ensures that the client returns a specific error for
// a certificate issued by an unknown authority.
func Test_CertificateSignedByUnknownAuthority(t *testing.T) {
	// We use httptest.NewTLSServer here which uses it's own invalid certificate.
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	config := &Configuration{
		Endpoint: mockServer.URL,
	}

	clientWrapper, err := New(config)
	if err != nil {
		t.Error(err)
	}

	_, err = clientWrapper.GetClusters(nil)
	if !clienterror.IsCertificateSignedByUnknownAuthorityError(err) {
		t.Errorf("Expected x509.UnknownAuthorityError, got %#v", err)
	}
}

func TestGetApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
		{
		  "metadata": {
			"name": "my-awesome-prometheus"
		  },
		  "spec": {
			"catalog": "sample-catalog",
			"name": "prometheus-chart",
			"namespace": "giantswarm",
			"version": "0.2.0",
			"user_config": {
			  "configmap": {
				"name": "prometheus-user-values",
				"namespace": "123ab"
			  }
			}
		  },
		  "status": {
			"app_version": "1.0.0",
			"release": {
			  "last_deployed": "2019-04-08T12:34:00Z",
			  "status": "DEPLOYED"
			},
			"version": "0.2.0"
		  }
		}
	  ]`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	app, err := gsClient.GetApp("cluster-id", "my-awesome-prometheus", nil)

	if err != nil {
		t.Error(err)
	}

	if app.Metadata.Name != "my-awesome-prometheus" {
		t.Errorf("Expected  name should be my-awesome-prometheus got %v", app.Metadata.Name)
	}
	if app.Status.AppVersion != "1.0.0" {
		t.Errorf("Expected app version should be 1.0.0 got %v", app.Status.Version)
	}
}

func TestGetAppStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
		{
		  "metadata": {
			"name": "my-awesome-prometheus"
		  },
		  "spec": {
			"catalog": "sample-catalog",
			"name": "prometheus-chart",
			"namespace": "giantswarm",
			"version": "0.2.0",
			"user_config": {
			  "configmap": {
				"name": "prometheus-user-values",
				"namespace": "123ab"
			  }
			}
		  },
		  "status": {
			"app_version": "1.0.0",
			"release": {
			  "last_deployed": "2019-04-08T12:34:00Z",
			  "status": "DEPLOYED"
			},
			"version": "0.2.0"
		  }
		}
	  ]`))
	}))
	defer ts.Close()

	config := &Configuration{
		Endpoint: ts.URL,
	}

	gsClient, err := New(config)
	if err != nil {
		t.Error(err)
	}

	appStatus, err := gsClient.GetAppStatus("cluster-id", "my-awesome-prometheus", nil)

	if err != nil {
		t.Error(err)
	}

	if appStatus != "DEPLOYED" {
		t.Errorf("Expected status should be DEPLOYED got %v", appStatus)
	}
}
