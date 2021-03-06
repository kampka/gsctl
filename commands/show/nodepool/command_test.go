package nodepool

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/giantswarm/gsctl/testutils"
)

// Test_ShowNodePool tries cases where we don't expect any errors.
func Test_ShowNodePool(t *testing.T) {
	var testCases = []struct {
		name         string
		responseBody string
	}{
		{
			name: "aws",
			responseBody: `{
				"id": "nodepool-id",
				"name": "Application servers",
				"availability_zones": ["eu-west-1a", "eu-west-1c"],
				"scaling": {"min": 3, "max": 10},
				"node_spec": {"aws": {"instance_type": "c5.large", "use_alike_instance_types": false, "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}}, "volume_sizes_gb": {"docker": 100, "kubelet": 100}},
				"status": {"nodes": 3, "nodes_ready": 3},
				"subnet": "10.1.0.0/24"
			}`,
		},
		{
			name: "aws, instance type non-existing",
			responseBody: `{
				"id":"nodepool-id",
				"name":"awesome-nodepool",
				"availability_zones":["europe-west-1b","europe-central-1a","europe-central-1b"],
				"scaling":{"min":2,"max":5},
				"node_spec":{"aws":{"instance_type":"nonexisting", "use_alike_instance_types": false, "instance_distribution": {"on_demand_base_capacity": 0, "on_demand_percentage_above_base_capacity": 0}},"labels":["web-compute"],"volume_sizes":{"docker":100,"kubelet":100}},
				"status":{"nodes":4,"nodes_ready":4},
				"subnet":"10.1.0.0/24"
			}`,
		},
		{
			name: "azure",
			responseBody: `{
				"id": "nodepool-id",
				"name": "Application servers",
				"availability_zones": ["1", "2"],
				"node_spec": {"azure": {"vm_size": "Standard_D4s_v3"}},
                "scaling":{"min":2,"max":5},
				"status": {"nodes": 3, "nodes_ready": 3}
			}`,
		},
		{
			name: "azure, non-existent vm size",
			responseBody: `{
				"id": "nodepool-id",
				"name": "Application servers",
				"availability_zones": ["1", "2"],
				"node_spec": {"azure": {"vm_size": "weird_one"}},
				"scaling":{"min":2,"max":5},
				"status": {"nodes": 3, "nodes_ready": 3}
			}`,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == "GET" {
					switch uri := r.URL.Path; uri {
					case "/v4/clusters/":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`[
						{
							"id": "cluster-id",
							"name": "Name of the cluster",
							"owner": "acme"
						}
					]`))

					case "/v5/clusters/cluster-id/nodepools/nodepool-id/":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(tc.responseBody))

					default:
						t.Errorf("Unsupported route %s called in mock server", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"code": "RESOURCE_NOT_FOUND", "message": "Status for this cluster is not yet available."}`))
					}
				}
			}))
			defer mockServer.Close()

			// temp config
			configYAML := `last_version_check: 0001-01-01T00:00:00Z
updated: 2017-09-29T11:23:15+02:00
endpoints:
  ` + mockServer.URL + `:
    email: email@example.com
    token: some-token
selected_endpoint: ` + mockServer.URL
			fs := afero.NewMemMapFs()
			_, err := testutils.TempConfig(fs, configYAML)
			if err != nil {
				t.Error(err)
			}

			args := &Arguments{
				apiEndpoint:     mockServer.URL,
				authToken:       "some-token",
				clusterNameOrID: "cluster-id",
				nodePoolID:      "nodepool-id",
			}
			positionalArgs := []string{"cluster-id/nodepool-id"}

			nodePool, err := fetchNodePool(args)
			if err != nil {
				t.Errorf("Case %d: unexpected error '%s'", i, err)
			}

			if nodePool == nil {
				t.Fatalf("Case %d: Got Got nil instead of node pool details", i)
			}

			if nodePool.ID != "nodepool-id" {
				t.Errorf("Case %d: Got unexpected node pool ID %s", i, nodePool.ID)
			}

			// Execute our print function and check for errors.
			output := testutils.CaptureOutput(func() {
				ShowNodepoolCommand.SetArgs([]string{"--endpoint", mockServer.URL, "--token", "my-token"})
				printResult(ShowNodepoolCommand, positionalArgs)
			})
			if strings.Contains(output, "Error") {
				t.Errorf("Case %d: Contained 'Error'", i)
			}
		})
	}

}
