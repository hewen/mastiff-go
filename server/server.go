// Package server package provides a collection of server implementations
package server

import (
	"log"
	"os"
	"os/signal"
	"sync"

	// automatically sets GOMAXPROCS to match the Linux container CPU quota.
	_ "go.uber.org/automaxprocs"
)

var (
	// stopFunc holds functions to be called during graceful shutdown.
	stopFuncMu       sync.Mutex
	stopFunc         []func()
	gracefulStopOnce sync.Once
)

// Server is an interface that defines methods for starting and stopping a server.
type Server interface {
	Start()
	Stop()
}

// Servers is a collection of Server instances.
type Servers struct {
	s []Server
}

// Add adds a server to the list of servers.
func (s *Servers) Add(server Server) {
	s.s = append(s.s, server)
}

// Start starts all registered servers.
func (s *Servers) Start() {
	var group sync.WaitGroup
	group.Add(len(s.s))
	AddGracefulStop(s.Stop)
	gracefulStop()

	for i := range s.s {
		go func(i int) {
			defer group.Done()
			s.s[i].Start()
		}(i)
	}

	group.Wait()
}

// Stop stops all registered servers.
func (s *Servers) Stop() {
	for i := range s.s {
		s.s[i].Stop()
	}
}

// AddGracefulStop registers a function to be called during graceful shutdown.
func AddGracefulStop(fn func()) {
	stopFuncMu.Lock()
	defer stopFuncMu.Unlock()

	stopFunc = append(stopFunc, fn)
}

// gracefulStop listens for an interrupt signal and executes registered stop functions.
func gracefulStop() {
	gracefulStopOnce.Do(func() {
		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt)
			<-sigint
			shutdown()
		}()
	})
}

func shutdown() {
	log.Println("shutdown service.")

	stopFuncMu.Lock()
	funcs := make([]func(), len(stopFunc))
	copy(funcs, stopFunc)
	stopFuncMu.Unlock()

	for i := range funcs {
		if funcs[i] == nil {
			log.Printf("stopFunc[%d] is nil", i)
			continue
		}
		funcs[i]()
	}
}
