package api

//go:generate go run golang.org/x/tools/cmd/stringer -type=TurnErrorCode -linecomment -trimprefix=TurnErrorCode
//go:generate go run golang.org/x/tools/cmd/stringer -type=TurnStatus -linecomment --trimprefix TurnStatus

type TurnStatus int

const (
	TurnStatusInit TurnStatus = iota
	TurnStatusProcessing
	TurnStatusSuccess
	TurnStatusFailed
)

type ErrorCode int

const (
	ErrorCodeConversationHasInitTurn ErrorCode = 1000
)

type TurnErrorCode int

const (
	_                                 TurnErrorCode = iota
	TurnErrorCodeInternalServer                     // Internal Server Error
	TurnErrorCodeConvNotFound                       // Conversation Not Found
	TurnErrorCodeBotNotFound                        // Bot Not Found
	TurnErrorCodeMiddlewareError                    // Middleware Error
	TurnErrorCodeRenderPromptError                  // Render Prompt Error
	TurnErrorCodeChatModelNotFound                  // Chat Model Not Found
	TurnErrorCodeChatModelCallTimeout               // Chat Model Call Timeout
	TurnErrorCodeChatModelCallError                 // Chat Model Call Error
)

type MiddlewareErrorCode int

const (
	_ MiddlewareErrorCode = iota
	MiddlewareErrorCodeConfigInvalid
	MiddlewareErrorCodeProcessFailed
	MiddlewareErrorCodeTimeout
)
