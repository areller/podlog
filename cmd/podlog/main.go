package main

import (
	"os/signal"
	"strings"
	"fmt"
    "github.com/sirupsen/logrus"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    //"k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/api/core/v1"
    "os"
)

func buildClient() *kubernetes.Clientset {
    var config *rest.Config
    var err error
    if val, ok := os.LookupEnv("IS_REMOTE"); ok && strings.ToLower(val) == "true" {
        // For debugging purposes
        config = &rest.Config{
            Host: os.Getenv("REST_HOST"),
            TLSClientConfig: rest.TLSClientConfig{
                CertFile: GetEnvVarPath("REST_CERT_FILE"),
                KeyFile: GetEnvVarPath("REST_KEY_FILE"),
                CAFile: GetEnvVarPath("REST_CA_FILE"),
            },
        }
    } else {
        // Real life
        config, err = rest.InClusterConfig()
        if err != nil {
            logrus.WithError(err).Panic()
        }
    }

    client, err := kubernetes.NewForConfig(config)
    if err != nil {
        logrus.WithError(err).Panic()
    }

    return client
}

func main() {
    sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

    client := buildClient()
    podRegistry := NewPodRegistry(client)
    logExtractor := NewLogExtractor(client)

    go func() {
        err := podRegistry.Run()
        if err != nil {
            logrus.WithError(err).Panic()
        }
    }()

    podEventsListener := podRegistry.Events.Listen()
    for {
        select {
        case ev := <-podEventsListener.Output():
            {
                podEv := ev.(PodEvent)
                switch podEv.Type {
                case Added:
                    logExtractor.AddPod(podEv.Pod)
                case Removed:
                    logExtractor.RemovePod(podEv.Pod)
                }
            }
        case <-sigChan:
            {
                logrus.Info("Closing Everything")
                podRegistry.Close()
                logExtractor.Close()
                return
            }
        }
    }
}