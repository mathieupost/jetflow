package memory

import (
	"context"
	"log"
	"sync"

	"github.com/mathieupost/jetflow"
)

var _ jetflow.Dispatcher = (*Dispatcher)(nil)

type Dispatcher struct {
	outbox           chan jetflow.Request
	inbox            chan jetflow.Response
	responseChannels sync.Map
}

func NewDispatcher(outbox chan jetflow.Request, inbox chan jetflow.Response) *Dispatcher {
	d := &Dispatcher{
		outbox:           outbox,
		inbox:            inbox,
		responseChannels: sync.Map{},
	}
	go d.processResponses()
	return d
}

func (d *Dispatcher) Dispatch(ctx context.Context, call jetflow.Request) (chan jetflow.Response, error) {
	// Setup the channel to which the response will be sent.
	responseChan := make(chan jetflow.Response, 10)
	d.responseChannels.Store(call.RequestID, responseChan)

	d.outbox <- call

	return responseChan, nil
}

func (d *Dispatcher) processResponses() {
	defer func() {
		log.Fatalln("Dispatcher processResponses exited")
	}()
	for response := range d.inbox {
		c, ok := d.responseChannels.LoadAndDelete(response.RequestID)
		if !ok {
			log.Fatalln("response not found", response)
		}

		responseChan := c.(chan jetflow.Response)
		responseChan <- response
	}
}
