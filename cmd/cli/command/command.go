package command

// Basic command interface
type ICommand interface {
	Execute() error
}
