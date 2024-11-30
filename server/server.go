package server

import (
	"log"
	"os"
	"os/signal"
	"sync"

	// automatically sets GOMAXPROCS to match the Linux container CPU quota.
	_ "go.uber.org/automaxprocs"
)

var stopFunc []func()

type Server interface {
	Start()
	Stop()
}

type Servers struct {
	s []Server
}

func (s *Servers) Add(server Server) {
	s.s = append(s.s, server)
}

func (s *Servers) Start() {
	var group sync.WaitGroup
	group.Add(len(s.s))
	AddGracefulStop(s.Stop)

	for i := range s.s {
		go func(i int) {
			defer group.Done()
			s.s[i].Start()
		}(i)
	}

	group.Wait()
}

func (s *Servers) Stop() {
	for i := range s.s {
		s.s[i].Stop()
	}
}

func AddGracefulStop(fn func()) {
	stopFunc = append(stopFunc, fn)
}

func gracefulStop() {
	waitClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("shutdown service.")
		for i := range stopFunc {
			stopFunc[i]()
		}
		close(waitClosed)
	}()
}
