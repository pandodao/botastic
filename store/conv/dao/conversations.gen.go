// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gen/helper"

	"github.com/pandodao/botastic/core"
)

func newConversation(db *gorm.DB, opts ...gen.DOOption) conversation {
	_conversation := conversation{}

	_conversation.conversationDo.UseDB(db, opts...)
	_conversation.conversationDo.UseModel(&core.Conversation{})

	tableName := _conversation.conversationDo.TableName()
	_conversation.ALL = field.NewAsterisk(tableName)
	_conversation.ID = field.NewString(tableName, "id")
	_conversation.Lang = field.NewString(tableName, "lang")
	_conversation.UserIdentity = field.NewString(tableName, "user_identity")
	_conversation.BotID = field.NewUint64(tableName, "bot_id")
	_conversation.AppID = field.NewUint64(tableName, "app_id")
	_conversation.CreatedAt = field.NewTime(tableName, "created_at")
	_conversation.UpdatedAt = field.NewTime(tableName, "updated_at")
	_conversation.DeletedAt = field.NewTime(tableName, "deleted_at")

	_conversation.fillFieldMap()

	return _conversation
}

type conversation struct {
	conversationDo

	ALL          field.Asterisk
	ID           field.String
	Lang         field.String
	UserIdentity field.String
	BotID        field.Uint64
	AppID        field.Uint64
	CreatedAt    field.Time
	UpdatedAt    field.Time
	DeletedAt    field.Time

	fieldMap map[string]field.Expr
}

func (c conversation) Table(newTableName string) *conversation {
	c.conversationDo.UseTable(newTableName)
	return c.updateTableName(newTableName)
}

func (c conversation) As(alias string) *conversation {
	c.conversationDo.DO = *(c.conversationDo.As(alias).(*gen.DO))
	return c.updateTableName(alias)
}

func (c *conversation) updateTableName(table string) *conversation {
	c.ALL = field.NewAsterisk(table)
	c.ID = field.NewString(table, "id")
	c.Lang = field.NewString(table, "lang")
	c.UserIdentity = field.NewString(table, "user_identity")
	c.BotID = field.NewUint64(table, "bot_id")
	c.AppID = field.NewUint64(table, "app_id")
	c.CreatedAt = field.NewTime(table, "created_at")
	c.UpdatedAt = field.NewTime(table, "updated_at")
	c.DeletedAt = field.NewTime(table, "deleted_at")

	c.fillFieldMap()

	return c
}

func (c *conversation) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := c.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (c *conversation) fillFieldMap() {
	c.fieldMap = make(map[string]field.Expr, 8)
	c.fieldMap["id"] = c.ID
	c.fieldMap["lang"] = c.Lang
	c.fieldMap["user_identity"] = c.UserIdentity
	c.fieldMap["bot_id"] = c.BotID
	c.fieldMap["app_id"] = c.AppID
	c.fieldMap["created_at"] = c.CreatedAt
	c.fieldMap["updated_at"] = c.UpdatedAt
	c.fieldMap["deleted_at"] = c.DeletedAt
}

func (c conversation) clone(db *gorm.DB) conversation {
	c.conversationDo.ReplaceConnPool(db.Statement.ConnPool)
	return c
}

func (c conversation) replaceDB(db *gorm.DB) conversation {
	c.conversationDo.ReplaceDB(db)
	return c
}

type conversationDo struct{ gen.DO }

type IConversationDo interface {
	WithContext(ctx context.Context) IConversationDo

	CreateConversation(ctx context.Context, conv *core.Conversation) (err error)
	GetConversation(ctx context.Context, id string) (result *core.Conversation, err error)
	GetConvTurnsByConversationID(ctx context.Context, conversationID string, limit int) (result []*core.ConvTurn, err error)
	CreateConvTurn(ctx context.Context, convID string, botID uint64, appID uint64, userID uint64, uid string, request string, bo core.BotOverride) (result uint64, err error)
	GetConvTurns(ctx context.Context, ids []uint64) (result []*core.ConvTurn, err error)
	GetConvTurn(ctx context.Context, id uint64) (result *core.ConvTurn, err error)
	GetConvTurnsByStatus(ctx context.Context, excludeIDs []uint64, status []int) (result []*core.ConvTurn, err error)
	UpdateConvTurn(ctx context.Context, id uint64, response string, promptTokens int64, completionTokens int64, totalTokens int64, status int, mr core.MiddlewareResults, tpe *core.TurnProcessError) (err error)
}

// INSERT INTO "conversations"
// (
//
//	id, lang, user_identity, bot_id, app_id, created_at, updated_at
//
// ) VALUES (
//
//	@conv.ID, @conv.Lang, @conv.UserIdentity, @conv.BotID, @conv.AppID, NOW(), NOW()
//
// )
func (c conversationDo) CreateConversation(ctx context.Context, conv *core.Conversation) (err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, conv.ID)
	params = append(params, conv.Lang)
	params = append(params, conv.UserIdentity)
	params = append(params, conv.BotID)
	params = append(params, conv.AppID)
	generateSQL.WriteString("INSERT INTO \"conversations\" ( id, lang, user_identity, bot_id, app_id, created_at, updated_at ) VALUES ( ?, ?, ?, ?, ?, NOW(), NOW() ) ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Exec(generateSQL.String(), params...) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT * FROM "conversations" WHERE id = @id AND deleted_at IS NULL
func (c conversationDo) GetConversation(ctx context.Context, id string) (result *core.Conversation, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("SELECT * FROM \"conversations\" WHERE id = ? AND deleted_at IS NULL ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT * FROM "conv_turns" WHERE conversation_id = @conversationID ORDER BY id DESC LIMIT @limit
func (c conversationDo) GetConvTurnsByConversationID(ctx context.Context, conversationID string, limit int) (result []*core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, conversationID)
	params = append(params, limit)
	generateSQL.WriteString("SELECT * FROM \"conv_turns\" WHERE conversation_id = ? ORDER BY id DESC LIMIT ? ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Find(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// INSERT INTO "conv_turns"
//
//	(
//	"conversation_id", "bot_id", "app_id", "user_id",
//
// "user_identity",
// "request", "response", "status", "bot_override",
// "created_at", "updated_at"
//
//	)
//
// VALUES
//
//		(
//	 @convID, @botID, @appID, @userID,
//	 @uid,
//	 @request, '', 0, @bo,
//	 NOW(), NOW()
//
// )
// RETURNING "id"
func (c conversationDo) CreateConvTurn(ctx context.Context, convID string, botID uint64, appID uint64, userID uint64, uid string, request string, bo core.BotOverride) (result uint64, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, convID)
	params = append(params, botID)
	params = append(params, appID)
	params = append(params, userID)
	params = append(params, uid)
	params = append(params, request)
	params = append(params, bo)
	generateSQL.WriteString("INSERT INTO \"conv_turns\" ( \"conversation_id\", \"bot_id\", \"app_id\", \"user_id\", \"user_identity\", \"request\", \"response\", \"status\", \"bot_override\", \"created_at\", \"updated_at\" ) VALUES ( ?, ?, ?, ?, ?, ?, '', 0, ?, NOW(), NOW() ) RETURNING \"id\" ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT *
// FROM "conv_turns" WHERE
// "id" IN (@ids)
func (c conversationDo) GetConvTurns(ctx context.Context, ids []uint64) (result []*core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, ids)
	generateSQL.WriteString("SELECT * FROM \"conv_turns\" WHERE \"id\" IN (?) ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Find(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT *
// FROM "conv_turns" WHERE
// "id" = @id
func (c conversationDo) GetConvTurn(ctx context.Context, id uint64) (result *core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, id)
	generateSQL.WriteString("SELECT * FROM \"conv_turns\" WHERE \"id\" = ? ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT *
// FROM "conv_turns"
// {{where}}
// "status" IN (@status)
//
//	{{if len(excludeIDs)>0}}
//	  AND "id" NOT IN (@excludeIDs)
//	{{end}}
//
// {{end}}
func (c conversationDo) GetConvTurnsByStatus(ctx context.Context, excludeIDs []uint64, status []int) (result []*core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	generateSQL.WriteString("SELECT * FROM \"conv_turns\" ")
	var whereSQL0 strings.Builder
	params = append(params, status)
	whereSQL0.WriteString("\"status\" IN (?) ")
	if len(excludeIDs) > 0 {
		params = append(params, excludeIDs)
		whereSQL0.WriteString("AND \"id\" NOT IN (?) ")
	}
	helper.JoinWhereBuilder(&generateSQL, whereSQL0)

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Find(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// UPDATE "conv_turns"
//
//		{{set}}
//			"response"=@response,
//	  "prompt_tokens"=@promptTokens,
//	  "completion_tokens"=@completionTokens,
//			"total_tokens"=@totalTokens,
//			"status"=@status,
//
// {{if mr != nil}}
//
//	"middleware_results"=@mr,
//
// {{end}}
// {{if tpe != nil}}
//
//	"error"=@tpe,
//
// {{end}}
//
//		"updated_at"=NOW()
//	{{end}}
//
// WHERE
//
//	"id"=@id
func (c conversationDo) UpdateConvTurn(ctx context.Context, id uint64, response string, promptTokens int64, completionTokens int64, totalTokens int64, status int, mr core.MiddlewareResults, tpe *core.TurnProcessError) (err error) {
	var params []interface{}

	var generateSQL strings.Builder
	generateSQL.WriteString("UPDATE \"conv_turns\" ")
	var setSQL0 strings.Builder
	params = append(params, response)
	params = append(params, promptTokens)
	params = append(params, completionTokens)
	params = append(params, totalTokens)
	params = append(params, status)
	setSQL0.WriteString("\"response\"=?, \"prompt_tokens\"=?, \"completion_tokens\"=?, \"total_tokens\"=?, \"status\"=?, ")
	if mr != nil {
		params = append(params, mr)
		setSQL0.WriteString("\"middleware_results\"=?, ")
	}
	if tpe != nil {
		params = append(params, tpe)
		setSQL0.WriteString("\"error\"=?, ")
	}
	setSQL0.WriteString("\"updated_at\"=NOW() ")
	helper.JoinSetBuilder(&generateSQL, setSQL0)
	params = append(params, id)
	generateSQL.WriteString("WHERE \"id\"=? ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Exec(generateSQL.String(), params...) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (c conversationDo) WithContext(ctx context.Context) IConversationDo {
	return c.withDO(c.DO.WithContext(ctx))
}

func (c *conversationDo) withDO(do gen.Dao) *conversationDo {
	c.DO = *do.(*gen.DO)
	return c
}
