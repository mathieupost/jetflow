package gen

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
	"github.com/mathieupost/jetflow/example/types"
)

var _ jetflow.OperatorHandler = (*UserHandler)(nil)

type UserHandler struct {
	user types.User
}

func NewUserHandler(id string) *UserHandler {
	user := types.NewUser(id)
	return &UserHandler{user}
}

func (o *UserHandler) Call(ctx context.Context, msg jetflow.OperatorCall) jetflow.Result {
	switch msg.Method {

	case "TransferBalance":
		var params Params_User_TransferBalance
		err := json.Unmarshal(msg.Params, &params)
		if err != nil {
			return jetflow.Result{
				Error: errors.Wrap(err, "unmarshal params"),
			}
		}
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
