package main

import (
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// check if we should report all events
var reportAll = os.Getenv("REPORT_ALL") == "true"

func main() {
	// get dsn
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		panic("missing SENTRY_DSN")
	}

	// initialize sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:        dsn,
		Debug:      os.Getenv("SENTRY_DEBUG") == "true",
		ServerName: "sentinel",
		Integrations: func([]sentry.Integration) []sentry.Integration {
			// disable all integrations
			return nil
		},
	})
	if err != nil {
		panic(err)
	}

	// get kube master and config
	kubeMaster := os.Getenv("KUBE_MASTER")
	kubeConfig := os.Getenv("KUBE_CONFIG")

	// prepare config
	var config *rest.Config

	// check kube master and config
	if kubeMaster != "" || kubeConfig != "" {
		// use provided kube master and config
		config, err = clientcmd.BuildConfigFromFlags(kubeMaster, kubeConfig)
		if err != nil {
			panic(err)
		}
	} else {
		// otherwise get in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	// get namespace
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = api.NamespaceAll
	}

	// create client set
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create list watch
	listWatch := cache.NewListWatchFromClient(
		clientSet.CoreV1().RESTClient(),
		"events",
		namespace,
		fields.Everything(),
	)

	// create informer controller
	_, controller := cache.NewInformer(
		listWatch,
		&api.Event{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				process(obj.(*api.Event))
			},
			UpdateFunc: func(_, obj interface{}) {
				process(obj.(*api.Event))
			},
		},
	)

	// run controller
	controller.Run(nil)
}

func process(event *api.Event) {
	// ignore normal events if report all is not set
	if event.Type == api.EventTypeNormal && !reportAll {
		return
	}

	// prepare level
	level := sentry.LevelInfo
	if event.Type == api.EventTypeWarning {
		level = sentry.LevelWarning
	}

	// prepare message
	message := fmt.Sprintf(
		"[%s] %s/%s: %s",
		event.InvolvedObject.Kind,
		event.InvolvedObject.Namespace,
		event.InvolvedObject.Name,
		event.Message,
	)

	// prepare sentry event
	sentryEvent := &sentry.Event{
		Message: message,
		Level:   level,
		Tags: map[string]string{
			"type":      event.Type,
			"reason":    event.Reason,
			"kind":      event.InvolvedObject.Kind,
			"name":      event.InvolvedObject.Name,
			"namespace": event.InvolvedObject.Namespace,
		},
		Extra: map[string]interface{}{
			"event":  event.Name,
			"count":  event.Count,
			"source": event.Source.Component,
		},
	}

	// capture event
	sentry.CaptureEvent(sentryEvent)

	// log info
	fmt.Printf("sent event: %s\n", message)
}
