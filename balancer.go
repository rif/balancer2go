// Simple balancer
package balancer2go

import (
	"log"
	"sync"
)

type Balancer struct {
	sync.RWMutex
	clients         map[string]Worker
	balancerChannel chan Worker
}

type Worker interface {
	Call(serviceMethod string, args interface{}, reply interface{}) error
	Close() error
}

/*
Constructor for RateList holding one slice for addreses and one slice for connections.
*/
func NewBalancer() *Balancer {
	r := &Balancer{clients: make(map[string]Worker), balancerChannel: make(chan Worker)} // leaving both slices to nil
	go func() {
		for {
			if len(r.clients) > 0 {
				for _, c := range r.clients {
					r.balancerChannel <- c
				}
			} else {
				r.balancerChannel <- nil
			}
		}
	}()
	return r
}

/*
Adds a client to the two  internal slices.
*/
func (bl *Balancer) AddClient(address string, client Worker) {
	bl.Lock()
	defer bl.Unlock()
	bl.clients[address] = client
	return
}

/*
Removes a client from the slices locking the readers and reseting the balancer index.
*/
func (bl *Balancer) RemoveClient(address string) {
	bl.Lock()
	defer bl.Unlock()
	delete(bl.clients, address)
	<-bl.balancerChannel
}

/*
Returns a client for the specifed address.
*/
func (bl *Balancer) GetClient(address string) (c Worker, exists bool) {
	bl.RLock()
	defer bl.RUnlock()
	c, exists = bl.clients[address]
	return
}

/*
Returns the next available connection at each call looping at the end of connections.
*/
func (bl *Balancer) Balance() (result Worker) {
	bl.RLock()
	defer bl.RUnlock()
	return <-bl.balancerChannel
}

func (bl *Balancer) Shutdown() {
	bl.Lock()
	defer bl.Unlock()
	var reply string
	for address, client := range bl.clients {
		client.Call("Responder.Shutdown", "", &reply)
		log.Printf("Shutdown rater %v: %v ", address, reply)
	}
}

func (bl *Balancer) GetClientAddresses() []string {
	bl.RLock()
	defer bl.RUnlock()
	var addresses []string
	for a, _ := range bl.clients {
		addresses = append(addresses, a)
	}
	return addresses
}
