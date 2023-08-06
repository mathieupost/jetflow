package jetstream

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
)

const (
	HEADER_KEY_MESSAGE_TYPE         = "MESSAGE_TYPE"
	HEADER_KEY_TRANSACTION_OPERATOR = "TRANSACTION_OPERATOR"

	MESSAGE_TYPE_OPERATOR_CALL = ""
	MESSAGE_TYPE_PREPARE       = "PREPARE"
	MESSAGE_TYPE_VOTE_YES      = "VOTE_YES"
	MESSAGE_TYPE_VOTE_NO       = "VOTE_NO"
	MESSAGE_TYPE_COMMIT        = "COMMIT"
	MESSAGE_TYPE_ROLLBACK      = "ROLLBACK"
)

type Consumer struct {
	id        uuid.UUID
	jetstream jetstream.JetStream
	client    *Client
	storage   jetflow.Storage
}

func NewConsumer(ctx context.Context, jetstream jetstream.JetStream, client *Client, storage jetflow.Storage) (*Consumer, error) {
	consumerID := uuid.New()

	consumer := &Consumer{
		id:        consumerID,
		jetstream: jetstream,
		client:    client,
		storage:   storage,
	}
	consumer.initConsumer(ctx)

	return consumer, nil
}

func (c *Consumer) initConsumer(ctx context.Context) error {
	log.Println("Consumer.initConsumer")
	consumer, err := c.jetstream.CreateOrUpdateConsumer(
		ctx,
		STREAM_NAME_OPERATOR,
		jetstream.ConsumerConfig{
			Durable:   "Consumer-" + c.id.String(),
			AckPolicy: jetstream.AckExplicitPolicy,
		},
	)
	if err != nil {
		return errors.Wrap(err, "create consumer")
	}
	log.Println("Consumer.initConsumer", consumer.CachedInfo().Name)

	_, err = consumer.Consume(func(msg jetstream.Msg) {
		msgType := msg.Headers().Get(HEADER_KEY_MESSAGE_TYPE)
		switch msgType {
		case MESSAGE_TYPE_OPERATOR_CALL:
			go c.handleOperatorCall(msg)
		case MESSAGE_TYPE_PREPARE:
			go c.handlePrepare(msg)
		case MESSAGE_TYPE_VOTE_YES:
			go c.handleVoteYes(msg)
		case MESSAGE_TYPE_VOTE_NO:
			go c.handleVoteNo(msg)
		case MESSAGE_TYPE_COMMIT:
			go c.handleCommit(msg)
		case MESSAGE_TYPE_ROLLBACK:
			go c.handleRollback(msg)
		}
	})
	if err != nil {
		return errors.Wrap(err, "execute consumer")
	}

	return nil
}

func (c *Consumer) handleOperatorCall(msg jetstream.Msg) {
	headers := msg.Headers()
	clientID := headers.Get(HEADER_KEY_CLIENT_ID)
	requestID := headers.Get(HEADER_KEY_REQUEST_ID)
	initialCallID := headers.Get(HEADER_KEY_INITIAL_CALL_ID)
	callID := headers.Get(HEADER_KEY_CALL_ID)
	log.Println("Consumer.handleOperatorCall", callID)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	ctx = ctxWithRequestID(ctx, requestID)
	ctx = ctxWithInitialCallID(ctx, initialCallID)
	ctx = ctxAddInvolvedOperators(ctx, msg.Subject())

	// Unmarshal the method and parameters.
	var call jetflow.OperatorCall
	err := json.Unmarshal(msg.Data(), &call)
	if err != nil {
		log.Fatalln(err.Error(), msg.Data())
	}
	msg.Ack()

	operatorID := call.ID
	operatorType := call.Type

	var retries int
	var success bool
	var result jetflow.Result
	for !success {
		log.Println("Consumer.handleOperatorCall retries:", retries)
		retries++

		operator := c.storage.Get(ctx, operatorType, operatorID, requestID)
		result = operator.Call(ctx, c.client, call)

		if callID == initialCallID {
			involvedOperators := involvedOperatorsFromCtx(ctx)
			c.prepareOperators(*involvedOperators)
			// TODO prepare and commit all operators involved in this request.
		}
		success = c.storage.Prepare(ctx, operatorType, operatorID, requestID)
		if !success {
			continue
		}

		success = c.storage.Commit(ctx, operatorType, operatorID, requestID)
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Fatalln(err.Error(), result)
	}

	// Send back to the caller.
	res := nats.NewMsg("CLIENT." + clientID)
	res.Header.Set(HEADER_KEY_CALL_ID, callID)
	involvedOperators := involvedOperatorsFromCtx(ctx)
	for _, operator := range *involvedOperators {
		res.Header.Add(HEADER_KEY_INVOLVED_OPERATORS, operator)
	}
	res.Data = data
	_, err = c.jetstream.PublishMsg(ctx, res)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if callID == initialCallID {
		log.Println("Consumer published msg", res.Subject, "involvedOperators", involvedOperators)
	}
}

func (c *Consumer) prepareOperators(involvedOperators []string) {
	for _, operator := range involvedOperators {
		prepareMsg := nats.NewMsg(operator)
		// TODO send prepare messages to operators.
		_ = prepareMsg
	}
}

func (c *Consumer) handlePrepare(msg jetstream.Msg) {
	// TODO handle prepare message and respond with a vote.
}

func (c *Consumer) handleVoteYes(msg jetstream.Msg) {
	// TODO handle vote yes message and respond with a commit if it was the last one.
}

func (c *Consumer) handleVoteNo(msg jetstream.Msg) {
	// TODO handle vote no message and respond with a rollback.
}

func (c *Consumer) handleCommit(msg jetstream.Msg) {
	// TODO commit the operator version.
}

func (c *Consumer) handleRollback(msg jetstream.Msg) {
	// TODO delete the operator version.
}
