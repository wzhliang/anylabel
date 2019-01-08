package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	key   string
	value string
)

func labelPodMutator(_ context.Context, obj metav1.Object) (bool, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return false, nil
	}

	// Mutate our object with the required annotations.
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	_, ok = pod.Labels[key]
	if !ok {
		pod.Labels["mutated"] = "true"
		pod.Labels[key] = value
	} else {
		fmt.Println("Cowardly refusing to overwrite current value.")
	}

	return false, nil
}

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")
	fl.StringVar(&key, "key", "", "Label key")
	fl.StringVar(&value, "value", "", "Label value")

	fl.Parse(os.Args[1:])
	return cfg
}

func main() {
	logger := &log.Std{Debug: true}

	cfg := initFlags()

	// Create our mutator
	mt := mutatingwh.MutatorFunc(labelPodMutator)

	mcfg := mutatingwh.WebhookConfig{
		Name: "podAnnotate",
		Obj:  &corev1.Pod{},
	}
	wh, err := mutatingwh.NewWebhook(mcfg, mt, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}
	logger.Infof("Listening on :8080")
	err = http.ListenAndServeTLS(":8080", cfg.certFile, cfg.keyFile, whHandler)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
