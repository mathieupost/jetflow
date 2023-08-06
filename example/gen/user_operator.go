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

func (o *UserHandler) Call(ctx context.Context, client jetflow.Client, msg jetflow.OperatorCall) jetflow.Result {
	log.Println("UserHandler.Call:", msg.ID, msg.Method, string(msg.Params), o.user)
	switch msg.Method {

	case "TransferBalance":
		var params Params_User_TransferBalance
		err := json.Unmarshal(msg.Params, &params)
		if err != nil {
			return jetflow.Result{
				Error: errors.Wrap(err, "unmarshal params"),
			}
		}
		params.U2.client = client
		err = o.user.TransferBalance(ctx, params.U2, params.Amount)
		return jetflow.Result{
			Error: err,
		}

	case "AddBalance":
		var params Params_User_AddBalance
		err := json.Unmarshal(msg.Params, &params)
		if err != nil {
			return jetflow.Result{
				Error: errors.Wrap(err, "unmarshal params"),
			}
		}
		err = o.user.AddBalance(ctx, params.Amount)
		return jetflow.Result{
			Error: err,
		}
	}

	return jetflow.Result{
		Error: errors.New("undefined method"),
	}
}
