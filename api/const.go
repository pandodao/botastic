package api

type TurnStatus int

const (
	TurnStatusInit TurnStatus = iota
	TurnStatusProcessing
	TurnStatusSuccess
	TurnStatusFailed
)
