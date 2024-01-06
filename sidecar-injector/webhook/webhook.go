package hook

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=sidecar-injector-v2.default.svc

// SidecarInjector annotates Pods
type SidecarInjector struct {
	Name    string
	Decoder *admission.Decoder
}

type Config struct {
	Containers []corev1.Container `yaml:"containers"`
}

func (si *SidecarInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	err := si.Decoder.Decode(req, pod)
	if err != nil {
		_ = logger.Log("Affinity-Injector: cannot decode", true)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	_ = logger.Log("Injecting affinity...", true)
	extendWithDispatcher(pod, logger)
	_ = logger.Log("Affinity  injected.", true)

	marshaledPod, err := json.Marshal(pod)

	if err != nil {
		_ = logger.Log("Affinity-Injector: cannot marshal", true)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func extendWithDispatcher(pod *corev1.Pod, logger log.Logger) {
	appLabel := pod.Labels["app"]
	if strings.TrimSpace(appLabel) == "" {
		_ = logger.Log("No 'app' label, skipping")

		return
	}

	_, ok := pod.Labels["affinity-inject"]
	if !ok {
		_ = logger.Log("No 'affinity-inject' label, skipping")

		return
	}

	labelsMap := map[string]string{
		"app": appLabel,
	}

	pod.Spec.TopologySpreadConstraints = append(pod.Spec.TopologySpreadConstraints, corev1.TopologySpreadConstraint{
		MaxSkew:           2,
		TopologyKey:       "topology.kubernetes.io/zone",
		WhenUnsatisfiable: "ScheduleAnyway",
		LabelSelector: &v1.LabelSelector{
			MatchLabels: labelsMap,
		},
	})

	if pod.Spec.Affinity != nil {
		return
	}

	pod.Spec.Affinity = new(corev1.Affinity)
	pod.Spec.Affinity.PodAntiAffinity = new(corev1.PodAntiAffinity)

	pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
		pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		corev1.WeightedPodAffinityTerm{
			PodAffinityTerm: corev1.PodAffinityTerm{
				LabelSelector: &v1.LabelSelector{
					MatchLabels: labelsMap,
				},
				TopologyKey: "kubernetes.io/hostname",
			},
			Weight: 100,
		},
	)
}
