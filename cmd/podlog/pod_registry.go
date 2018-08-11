package main

import (
	"sync"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"github.com/areller/multichan"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Added = iota
	Removed = iota
)

type PodEvent struct {
	Pod *v1.Pod
	Type int
}

type PodRegistry struct {
	client *kubernetes.Clientset
	closeChan chan struct{}
	closeAllWg sync.WaitGroup
	Events *multichan.Chan
}

func (pr *PodRegistry) handleEvent(ev watch.Event) {
	pod, ok := ev.Object.(*v1.Pod)
	if !ok {
		logrus.WithField("obj", ev.Object).Warn("Object from watch event is not a pod")
		return
	}

	switch ev.Type {
	case watch.Added:
		pr.Events.Input() <- PodEvent{
			Pod: pod,
			Type: Added,
		}
	case watch.Deleted:
		pr.Events.Input() <- PodEvent{
			Pod: pod,
			Type: Removed,
		}
	}
}

func (pr *PodRegistry) Close() {
	pr.Events.Close()
	pr.closeAllWg.Add(1)
	close(pr.closeChan)
	pr.closeAllWg.Wait()
}

func (pr *PodRegistry) Run() error {
	wi, err := pr.client.CoreV1().Pods("").Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for {
		select {
		case <-pr.closeChan:
			wi.Stop()
			pr.closeAllWg.Done()
			return nil
		case ev := <-wi.ResultChan():
			pr.handleEvent(ev)
		}
	}
}

func NewPodRegistry(client *kubernetes.Clientset) *PodRegistry {
	return &PodRegistry{
		client: client,
		closeChan: make(chan struct{}, 1),
		Events: multichan.New(),
	}
}