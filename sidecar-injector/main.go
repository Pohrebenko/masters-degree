package main

import (
	hook "github.com/krvarma/wh/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var log = logf.Log.WithName("example-controller")

func main() {
	logf.SetLogger(zap.New())
	entryLog := log.WithName("entrypoint")

	// Setup webhooks
	entryLog.Info("setting up webhook server")

	hndl, err := admission.StandaloneWebhook(
		&webhook.Admission{Handler: &hook.SidecarInjector{
			Name:    "NodeManager",
			Decoder: admission.NewDecoder(runtime.NewScheme()),
		}},
		admission.StandaloneOptions{
			MetricsPath: "",
		},
	)
	if err != nil {
		entryLog.Error(err, "StandaloneWebhook")
		os.Exit(1)
	}

	http.Handle("/mutate", hndl)

	http.ListenAndServeTLS(
		":8443",
		"/etc/webhook/certs/tls.crt",
		"/etc/webhook/certs/tls.key",
		nil,
	)
}
