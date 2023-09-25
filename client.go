package jetflow

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/mathieupost/jetflow/log"
)

var _ OperatorClient = (*Client)(nil)

type Client struct {
	mapping    ProxyFactoryMapping
	dispatcher Publisher
}

func NewClient(mapping ProxyFactoryMapping, dispatcher Publisher) *Client {
	return &Client{
		mapping:    mapping,
		dispatcher: dispatcher,
	}
}

func (c *Client) Call(ctx context.Context, call *Request) (res []byte, err error) {
	originalctx := ctx
	operatorspan, ok := ctx.Value("SPAN").(*trace.Span)
	if ok {
		(*operatorspan).End()
		defer func() {
			_, *operatorspan = otel.Tracer("").Start(originalctx, "operator.Handle")
		}()
	}

	ctx, span := otel.Tracer("client").Start(ctx, "jetflow.Client.Call."+call.Name+"."+call.Method)
	defer span.End()
	span.SetAttributes(
		attribute.String("operator", call.Name),
		attribute.String("id", call.ID),
	)

	spanID := span.SpanContext().SpanID().String()
	call.OperationID = OperationIDFromContext(ctx, spanID)
	call.RequestID = spanID

	log.Println("Client.Call:\n", call)

	replyChan, err := c.dispatcher.Publish(ctx, call)
	if err != nil {
		return nil, errors.Wrap(err, "dispatching request")
	}

	ctx, span = otel.Tracer("client").Start(ctx, "jetflow.Client.WaitForReply")
	defer span.End()
	select {
	case reply := <-replyChan:
		// Mutate the ctx's involved operators.
		for name, instances := range reply.InvolvedOperators {
			for id := range instances {
				ContextAddInvolvedOperator(ctx, name, id)
			}
		}
		return reply.Values, errors.Wrap(reply.Error, "dispatch call")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "dispatch call")
	}
}

func (c *Client) Find(ctx context.Context, id string, operator interface{}) error {
	value := reflect.ValueOf(operator)
	if value.Kind() != reflect.Pointer {
		return errors.New("operator must be a pointer")
	}

	value = value.Elem()
	if value.Kind() != reflect.Interface {
		return errors.New("operator must be an interface")
	}

	if value.Elem().Kind() != reflect.Invalid {
		return errors.New("operator already initalized")
	}

	name := value.Type().Name()
	factory, ok := c.mapping[name]
	if !ok {
		return errors.Errorf("operator '%s' not found", name)
	}
	operator = factory(id, c)
	value.Set(reflect.ValueOf(operator))

	return nil
}
