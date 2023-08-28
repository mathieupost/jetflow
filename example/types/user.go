package types

import (
	"context"

	"github.com/mathieupost/jetflow"
	"github.com/pkg/errors"
)

type User interface {
	jetflow.Operator
	TransferBalance(ctx context.Context, u2 User, amount int) error
	AddBalance(ctx context.Context, amount int) error
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

func (u *user) TransferBalance(ctx context.Context, u2 User, amount int) (err error) {
	if amount < 0 {
		return errors.New("amount cannot be negative")
	}

	if u.balance < amount {
		return errors.New("insufficient balance")
	}

	err = u2.AddBalance(ctx, amount)
	if err != nil {
		return errors.Wrap(err, "add balance to u2")
	}

	u.balance -= amount

	return nil
}

func (u *user) AddBalance(ctx context.Context, amount int) error {
	if amount < 0 {
		return errors.New("amount cannot be negative")
	}

	u.balance += amount

	return nil
}

func (u *user) GetBalance(ctx context.Context) (int, error) {
	return u.balance, nil
}
