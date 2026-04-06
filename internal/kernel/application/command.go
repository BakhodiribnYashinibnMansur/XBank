package application

import "context"

// Command is a marker interface for write operations (state mutations).
// Commands are named imperatively: CreateUser, DepositFunds, BlockCard.
type Command interface{}

// CommandHandler executes a command and returns a result.
type CommandHandler[C Command, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
