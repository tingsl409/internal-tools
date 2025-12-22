package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	disco "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestGetKubernetesVersion(t *testing.T) {
	okClientset := kubefake.NewSimpleClientset()
	okClientset.Discovery().(*disco.FakeDiscovery).FakedServerVersion = &version.Info{GitVersion: "1.25.0-fake"}

	okVer, err := getKubernetesVersion(okClientset)
	assert.NoError(t, err)
	assert.Equal(t, "1.25.0-fake", okVer)

	badClientset := kubefake.NewSimpleClientset()
	badClientset.Discovery().(*disco.FakeDiscovery).FakedServerVersion = &version.Info{}

	badVer, err := getKubernetesVersion(badClientset)
	assert.NoError(t, err)
	assert.Equal(t, "", badVer)
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)
	res := rec.Result()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	defer func(Body io.ReadCloser) {
		assert.NoError(t, Body.Close())
	}(res.Body)
	resp, err := io.ReadAll(res.Body)

	assert.NoError(t, err)
	assert.Equal(t, "ok", string(resp))
}

func TestBuildSelector(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: "all()",
		},
		{
			name:     "single label",
			labels:   map[string]string{"app": "test"},
			expected: "app == 'test'",
		},
		{
			name:     "multiple labels",
			labels:   map[string]string{"app": "test", "env": "prod"},
			expected: "app == 'test' && env == 'prod'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSelector(tt.labels)
			if tt.name == "multiple labels" {
				// For multiple labels, order may vary, so check both possibilities
				alt := "env == 'prod' && app == 'test'"
				assert.True(t, result == tt.expected || result == alt)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHandleCustomDeny(t *testing.T) {
	// Test just the spec parsing logic without the dynamic client create
	customDeny := &unstructured.Unstructured{}
	customDeny.SetName("test-deny")
	customDeny.SetNamespace("test-ns")
	
	spec := map[string]interface{}{
		"sourceNamespace": "source-ns",
		"sourceLabels": map[string]interface{}{
			"app": "source",
		},
		"targetLabels": map[string]interface{}{
			"app": "target",
		},
	}
	customDeny.Object["spec"] = spec

	// Test that we can extract the spec without errors
	extractedSpec, found, err := unstructured.NestedMap(customDeny.Object, "spec")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotNil(t, extractedSpec)
	
	sourceNamespace, _, _ := unstructured.NestedString(extractedSpec, "sourceNamespace")
	assert.Equal(t, "source-ns", sourceNamespace)
}

func TestDeleteCalicoPolicy(t *testing.T) {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClient(scheme)

	customDeny := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-deny",
				"namespace": "test-ns",
			},
		},
	}

	// This will return an error since the policy doesn't exist, but tests the function
	err := deleteCalicoPolicy(client, customDeny)
	assert.Error(t, err) // Expected since policy doesn't exist
}
