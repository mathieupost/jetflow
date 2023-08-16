package gen

import (
	"context"
	"encoding/json"
	"log"

	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
)

var _ jetflow.OperatorHandler = (*UserHandler)(nil)

type UserHandler struct {
	user types.User
}

func NewUserHandler(id string) jetflow.OperatorHandler {
	user := types.NewUser(id)
	return &UserHandler{user}
}

func (o *UserHandler) Handle(ctx context.Context, client jetflow.OperatorClient, call jetflow.Request) (res []byte, err error) {
	log.Println("UserHandler.Handle\n", call)
	switch call.Method {

	case "TransferBalance":
		var args User_TransferBalance_Args
		err := json.Unmarshal(call.Args, &args)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling TransferBalance args")
		}
		args.U2.client = client
		err = o.user.TransferBalance(ctx, args.U2, args.Amount)
		return nil, errors.Wrap(err, "calling user.TransferBalance")

	case "AddBalance":
		var args User_AddBalance_Args
		err := json.Unmarshal(call.Args, &args)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling AddBalance args")
		}
		err = o.user.AddBalance(ctx, args.Amount)
		return nil, errors.Wrap(err, "calling user.AddBalance")

	default:
		return nil, errors.Errorf("unknown method %s", call.Method)
	}
}
