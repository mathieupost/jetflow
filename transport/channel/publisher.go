package channel

import (
	"context"
	"log"
	"sync"

	"github.com/mathieupost/jetflow"
)

var _ jetflow.Publisher = (*Publisher)(nil)

type Publisher struct {
	outbox           chan jetflow.Request
	inbox            chan jetflow.Response
	responseChannels sync.Map
}

func NewPublisher(outbox chan jetflow.Request, inbox chan jetflow.Response) *Publisher {
	d := &Publisher{
		outbox:           outbox,
		inbox:            inbox,
		responseChannels: sync.Map{},
	}
	go d.processResponses()
	return d
}

func (d *Publisher) Publish(ctx context.Context, call jetflow.Request) (chan jetflow.Response, error) {
	// Setup the channel to which the response will be sent.
	responseChan := make(chan jetflow.Response, 10)
	d.responseChannels.Store(call.RequestID, responseChan)

	d.outbox <- call

	return responseChan, nil
}

func (d *Publisher) processResponses() {
	defer func() {
		log.Fatalln("Publisher processResponses exited")
	}()
	for response := range d.inbox {
		d.handleResponse(response)
	}
}

func (d *Publisher) handleResponse(response jetflow.Response) {
	c, ok := d.responseChannels.LoadAndDelete(response.RequestID)
	if !ok {
		log.Fatalln("response not found", response)
	}

	responseChan := c.(chan jetflow.Response)
	responseChan <- response
}
