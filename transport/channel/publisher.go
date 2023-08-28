package channel

import (
	"context"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/log"
)

var _ jetflow.Publisher = (*Publisher)(nil)

type Publisher struct {
	outbox           chan requestWithHeaders
	inbox            chan *jetflow.Response
	responseChannels sync.Map
}

type requestWithHeaders struct {
	*jetflow.Request
	headers http.Header
}

func NewPublisher() (*Publisher, chan requestWithHeaders, chan *jetflow.Response) {
	d := &Publisher{
		outbox:           make(chan requestWithHeaders),
		inbox:            make(chan *jetflow.Response),
		responseChannels: sync.Map{},
	}
	go d.processResponses()
	return d, d.outbox, d.inbox
}

func (d *Publisher) Publish(ctx context.Context, call *jetflow.Request) (chan *jetflow.Response, error) {
	ctx, span := otel.Tracer("").Start(ctx, "channel.Publisher.Publish")
	defer span.End()

	// Setup the channel to which the response will be sent.
	responseChan := make(chan *jetflow.Response)
	d.responseChannels.Store(call.RequestID, responseChan)

	req := requestWithHeaders{
		Request: call,
		headers: map[string][]string{},
	}

	// Inject the trace context into the message header.
	propagator := propagation.TraceContext{}
	carrier := propagation.HeaderCarrier(req.headers)
	propagator.Inject(ctx, carrier)

	d.outbox <- req

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

func (d *Publisher) handleResponse(response *jetflow.Response) {
	c, ok := d.responseChannels.LoadAndDelete(response.RequestID)
	if !ok {
		log.Fatalln("response not found", response)
	}

	responseChan := c.(chan *jetflow.Response)
	responseChan <- response
}
