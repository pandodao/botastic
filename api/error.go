package api

type ErrorCode int

const (
	ErrorCodeConversationHasInitTurn ErrorCode = 1000
)

type TurnErrorCode int

const (
	_                              TurnErrorCode = iota
	TurnErrorCodeInternalServer                  // Internal Server Error
	TurnErrorCodeConvNotFound                    // Conversation Not Found
	TurnErrorCodeBotNotFound                     // Bot Not Found
	TurnErrorCodeChatModelNotFound               // Chat Model Not Found
	TurnErrorCodeChatModelCallError
)

type TurnError struct {
	Code TurnErrorCode
	Msg  string
}

func (e *TurnError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}

	return e.Code.String()
}

func NewTurnError(code TurnErrorCode, msg ...string) *TurnError {
	e := &TurnError{
		Code: code,
	}
	if len(msg) > 0 {
		e.Msg = msg[0]
	}

	return e
}
