package types

import (
	"context"

	"github.com/mathieupost/jetflow"
	"github.com/pkg/errors"
)

type User interface {
	jetflow.Operator // Inherit the ID() string method of jetflow.Operator.
	TransferBalance(ctx context.Context, u2 User, amount int) (int, int, error)
	AddBalance(ctx context.Context, amount int) (int, error)
	GetBalance(ctx context.Context) (int, error)
}

func NewUser(id string) User {
	return &user{id: id, balance: 1000000}
}

var _ jetflow.Operator = (*user)(nil)

type user struct {
	id      string
	balance int
}

// ID implements jetflow.Operator interface.
func (u *user) ID() string {
	return u.id
}

func (u *user) TransferBalance(ctx context.Context, u2 User, amount int) (int, int, error) {
	if amount < 0 {
		return 0, 0, errors.New("amount cannot be negative")
	}

	if u.balance < amount {
		return 0, 0, errors.New("insufficient balance")
	}

	res, err := u2.AddBalance(ctx, amount)
	if err != nil {
		return 0, 0, errors.Wrap(err, "add balance to u2")
	}

	u.balance -= amount

	return u.balance, res, nil
}

func (u *user) AddBalance(ctx context.Context, amount int) (int, error) {
	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}

	u.balance += amount

	return u.balance, nil
}

func (u *user) GetBalance(ctx context.Context) (int, error) {
	return u.balance, nil
}
