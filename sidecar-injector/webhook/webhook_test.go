package hook

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// helper function to create a pod with specific labels
func createPodWithLabels(labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{},
	}
}

// TestExtendWithDispatcher tests the extendWithDispatcher function
func TestExtendWithDispatcher(t *testing.T) {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	// Test Cases
	tests := []struct {
		name                string
		podLabels           map[string]string
		expectedConstraints int
	}{
		{
			name:                "No Labels",
			podLabels:           map[string]string{},
			expectedConstraints: 0,
		},
		{
			name: "Only 'app' Label",
			podLabels: map[string]string{
				"app": "test-app",
			},
			expectedConstraints: 0,
		},
		{
			name: "Required Labels",
			podLabels: map[string]string{
				"app":             "test-app",
				"affinity-inject": "true",
			},
			expectedConstraints: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create Pod based on test case
			pod := createPodWithLabels(test.podLabels)

			// Call extendWithDispatcher
			extendWithDispatcher(pod, logger)

			// Assertions
			assert.Len(t, pod.Spec.TopologySpreadConstraints, test.expectedConstraints)

			if test.expectedConstraints > 0 {
				assert.NotNil(t, pod.Spec.Affinity)
				assert.NotNil(t, pod.Spec.Affinity.PodAntiAffinity)
				assert.NotEmpty(t, pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
			} else {
				assert.Nil(t, pod.Spec.Affinity)
			}
		})
	}
}
