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
	Code TurnErrorCode `json:"code"`
	Msg  string        `json:"msg"`
}
