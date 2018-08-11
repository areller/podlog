package main

import (
	"github.com/areller/multichan"
	"sync"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type LogExtractor struct {
	client *kubernetes.Clientset
	closeChans map[types.UID]chan struct{}
	closeAllWg sync.WaitGroup
}

func (le *LogExtractor) runForContainer(container v1.Container, pod *v1.Pod, closeChan *multichan.Chan, closeAllWg *sync.WaitGroup) {
	closeListener := closeChan.Listen()
}

func (le *LogExtractor) runForPod(pod *v1.Pod, closeChan chan struct{}) {
	var closeAllWg sync.WaitGroup
	closeAllContainers := multichan.New()
	for _, container := range pod.Spec.Containers {
		closeAllWg.Add(1)
		go le.runForContainer(container, pod, closeAllContainers, &closeAllWg)
	}

	for {
		select {
		case <-closeChan:
			{
				closeAllContainers.Close()
				closeAllWg.Wait()
				le.closeAllWg.Done()
				return
			}
		}
	}
}

func (le *LogExtractor) AddPod(pod *v1.Pod) {
	closeChan := make(chan struct{}, 1)
	le.closeChans[pod.UID] = closeChan
	go le.runForPod(pod, closeChan)
}

func (le *LogExtractor) RemovePod(pod *v1.Pod) {
	closeChan, ok := le.closeChans[pod.UID]
	if !ok {
		logrus.WithField("id", pod.UID).Error("Log extractor could not find pod with given UID")
		return
	}

	close(closeChan)
	delete(le.closeChans, pod.UID)
}

func (le *LogExtractor) Close() {
	for _, closeChan := range le.closeChans {
		le.closeAllWg.Add(1)
		close(closeChan)
	}

	le.closeAllWg.Wait()
}

func NewLogExtractor(client *kubernetes.Clientset) *LogExtractor {
	return &LogExtractor{
		client: client,
		closeChans: make(map[types.UID]chan struct{}),
	}
}