package api

type TurnStatus int

const (
	TurnStatusInit TurnStatus = iota
	TurnStatusSuccess
	TurnStatusFailed
)
