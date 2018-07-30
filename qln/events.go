package qln

import (
	"log"

	"github.com/mit-dci/lit/lnutil"
)

func (nd *LitNode) SendEvent(event lnutil.LitEvent) {
	nd.EventListenersMtx.Lock()
	log.Printf("Sending event to %d event listeners", len(nd.EventListeners))
	for _, l := range nd.EventListeners {
		select {
		case l <- event:
		default:
		}
	}
	nd.EventListenersMtx.Unlock()
}

func (nd *LitNode) RegisterEventListener(listener chan lnutil.LitEvent) {
	nd.EventListenersMtx.Lock()
	log.Printf("Adding event listener for RPC Events")
	nd.EventListeners = append(nd.EventListeners, listener)
	nd.EventListenersMtx.Unlock()
}

func (nd *LitNode) UnregisterEventListener(listener chan lnutil.LitEvent) {
	nd.EventListenersMtx.Lock()
	log.Printf("Removing event listener for RPC Events")

	for i, l := range nd.EventListeners {
		if l == listener {
			nd.EventListeners = append(nd.EventListeners[:i], nd.EventListeners[i+1:]...)
		}
	}
	nd.EventListenersMtx.Unlock()
}
