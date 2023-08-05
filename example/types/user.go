package types

import (
	"context"
	"log"

	"github.com/pkg/errors"

	"github.com/mathieupost/jetflow"
)

type User interface {
	jetflow.Operator
	TransferBalance(ctx context.Context, u2 User, amount int) error
	AddBalance(ctx context.Context, amount int) error
}

func NewUser(id string) User {
	return &user{id: id, balance: 10}
}

type user struct {
	id      string
	balance int
}

func (u *user) ID() string {
	return u.id
}

func (u *user) TransferBalance(ctx context.Context, u2 User, amount int) (err error) {
	log.Println("user(" + u.id + ").TransferBalance")
	if amount < 0 {
		return errors.New("negative amount")
	}

	if u.balance < amount {
		return errors.New("insufficient balance")
	}

	err = u2.AddBalance(ctx, amount)
	if err != nil {
		return errors.Wrap(err, "add balance to u2")
	}

	err = u.AddBalance(ctx, -amount)
	if err != nil {
		return errors.Wrap(err, "subtract balance from u")
	}

	return nil
}

func (u *user) AddBalance(ctx context.Context, amount int) error {
	log.Println("user(" + u.id + ").AddBalance")
	u.balance += amount

	return nil
}
