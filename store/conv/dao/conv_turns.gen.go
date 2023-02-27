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

func newConvTurn(db *gorm.DB, opts ...gen.DOOption) convTurn {
	_convTurn := convTurn{}

	_convTurn.convTurnDo.UseDB(db, opts...)
	_convTurn.convTurnDo.UseModel(&core.ConvTurn{})

	tableName := _convTurn.convTurnDo.TableName()
	_convTurn.ALL = field.NewAsterisk(tableName)
	_convTurn.ID = field.NewUint64(tableName, "id")
	_convTurn.ConversationID = field.NewString(tableName, "conversation_id")
	_convTurn.BotID = field.NewUint64(tableName, "bot_id")
	_convTurn.AppID = field.NewUint64(tableName, "app_id")
	_convTurn.UserIdentity = field.NewString(tableName, "user_identity")
	_convTurn.Request = field.NewString(tableName, "request")
	_convTurn.RequestToken = field.NewInt(tableName, "request_token")
	_convTurn.Response = field.NewString(tableName, "response")
	_convTurn.ResponseToken = field.NewInt(tableName, "response_token")
	_convTurn.Status = field.NewInt(tableName, "status")
	_convTurn.CreatedAt = field.NewTime(tableName, "created_at")
	_convTurn.UpdatedAt = field.NewTime(tableName, "updated_at")

	_convTurn.fillFieldMap()

	return _convTurn
}

type convTurn struct {
	convTurnDo

	ALL            field.Asterisk
	ID             field.Uint64
	ConversationID field.String
	BotID          field.Uint64
	AppID          field.Uint64
	UserIdentity   field.String
	Request        field.String
	RequestToken   field.Int
	Response       field.String
	ResponseToken  field.Int
	Status         field.Int
	CreatedAt      field.Time
	UpdatedAt      field.Time

	fieldMap map[string]field.Expr
}

func (c convTurn) Table(newTableName string) *convTurn {
	c.convTurnDo.UseTable(newTableName)
	return c.updateTableName(newTableName)
}

func (c convTurn) As(alias string) *convTurn {
	c.convTurnDo.DO = *(c.convTurnDo.As(alias).(*gen.DO))
	return c.updateTableName(alias)
}

func (c *convTurn) updateTableName(table string) *convTurn {
	c.ALL = field.NewAsterisk(table)
	c.ID = field.NewUint64(table, "id")
	c.ConversationID = field.NewString(table, "conversation_id")
	c.BotID = field.NewUint64(table, "bot_id")
	c.AppID = field.NewUint64(table, "app_id")
	c.UserIdentity = field.NewString(table, "user_identity")
	c.Request = field.NewString(table, "request")
	c.RequestToken = field.NewInt(table, "request_token")
	c.Response = field.NewString(table, "response")
	c.ResponseToken = field.NewInt(table, "response_token")
	c.Status = field.NewInt(table, "status")
	c.CreatedAt = field.NewTime(table, "created_at")
	c.UpdatedAt = field.NewTime(table, "updated_at")

	c.fillFieldMap()

	return c
}

func (c *convTurn) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := c.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (c *convTurn) fillFieldMap() {
	c.fieldMap = make(map[string]field.Expr, 12)
	c.fieldMap["id"] = c.ID
	c.fieldMap["conversation_id"] = c.ConversationID
	c.fieldMap["bot_id"] = c.BotID
	c.fieldMap["app_id"] = c.AppID
	c.fieldMap["user_identity"] = c.UserIdentity
	c.fieldMap["request"] = c.Request
	c.fieldMap["request_token"] = c.RequestToken
	c.fieldMap["response"] = c.Response
	c.fieldMap["response_token"] = c.ResponseToken
	c.fieldMap["status"] = c.Status
	c.fieldMap["created_at"] = c.CreatedAt
	c.fieldMap["updated_at"] = c.UpdatedAt
}

func (c convTurn) clone(db *gorm.DB) convTurn {
	c.convTurnDo.ReplaceConnPool(db.Statement.ConnPool)
	return c
}

func (c convTurn) replaceDB(db *gorm.DB) convTurn {
	c.convTurnDo.ReplaceDB(db)
	return c
}

type convTurnDo struct{ gen.DO }

type IConvTurnDo interface {
	WithContext(ctx context.Context) IConvTurnDo

	CreateConvTurn(ctx context.Context, convID string, botID uint64, appID uint64, uid string, request string, reqToken int) (result uint64, err error)
	GetConvTurns(ctx context.Context, ids []uint64) (result []*core.ConvTurn, err error)
	GetConvTurnsByStatus(ctx context.Context, status int) (result []*core.ConvTurn, err error)
	UpdateConvTurn(ctx context.Context, id uint64, response string, responseToken int, status int) (err error)
}

// INSERT INTO "conv_turns"
//
//	(
//	"conversation_id", "bot_id", "app_id", "user_identity",
//
// "request", "request_token", "response", "status",
// "created_at", "updated_at"
//
//	)
//
// VALUES
//
//		(
//	 @convID, @botID, @appID, @uid,
//	 @request, '', 0,
//	 NOW(), NOW()
//
// )
// RETURNING "id"
func (c convTurnDo) CreateConvTurn(ctx context.Context, convID string, botID uint64, appID uint64, uid string, request string, reqToken int) (result uint64, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, convID)
	params = append(params, botID)
	params = append(params, appID)
	params = append(params, uid)
	params = append(params, request)
	generateSQL.WriteString("INSERT INTO \"conv_turns\" ( \"conversation_id\", \"bot_id\", \"app_id\", \"user_identity\", \"request\", \"request_token\", \"response\", \"status\", \"created_at\", \"updated_at\" ) VALUES ( ?, ?, ?, ?, ?, '', 0, NOW(), NOW() ) RETURNING \"id\" ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT
//
//	"id", "conversation_id", "bot_id", "app_id", "user_identity",
//
// "request", "response", "status",
// "created_at", "updated_at"
// FROM "conv_turns" WHERE
// "id" IN (@ids)
func (c convTurnDo) GetConvTurns(ctx context.Context, ids []uint64) (result []*core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, ids)
	generateSQL.WriteString("SELECT \"id\", \"conversation_id\", \"bot_id\", \"app_id\", \"user_identity\", \"request\", \"response\", \"status\", \"created_at\", \"updated_at\" FROM \"conv_turns\" WHERE \"id\" IN (?) ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Find(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// SELECT
//
//	"id", "conversation_id", "bot_id", "app_id", "user_identity",
//
// "request", "response", "status",
// "created_at", "updated_at"
// FROM "conv_turns" WHERE
// "status"=@status
func (c convTurnDo) GetConvTurnsByStatus(ctx context.Context, status int) (result []*core.ConvTurn, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, status)
	generateSQL.WriteString("SELECT \"id\", \"conversation_id\", \"bot_id\", \"app_id\", \"user_identity\", \"request\", \"response\", \"status\", \"created_at\", \"updated_at\" FROM \"conv_turns\" WHERE \"status\"=? ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Raw(generateSQL.String(), params...).Find(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// UPDATE "conv_turns"
//
//	{{set}}
//		"response"=@response,
//		"response_token"=@responseToken,
//		"status"=@status,
//		"updated_at"=NOW()
//	{{end}}
//
// WHERE
//
//	"id"=@id
func (c convTurnDo) UpdateConvTurn(ctx context.Context, id uint64, response string, responseToken int, status int) (err error) {
	var params []interface{}

	var generateSQL strings.Builder
	generateSQL.WriteString("UPDATE \"conv_turns\" ")
	var setSQL0 strings.Builder
	params = append(params, response)
	params = append(params, responseToken)
	params = append(params, status)
	setSQL0.WriteString("\"response\"=?, \"response_token\"=?, \"status\"=?, \"updated_at\"=NOW() ")
	helper.JoinSetBuilder(&generateSQL, setSQL0)
	params = append(params, id)
	generateSQL.WriteString("WHERE \"id\"=? ")

	var executeSQL *gorm.DB
	executeSQL = c.UnderlyingDB().Exec(generateSQL.String(), params...) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (c convTurnDo) WithContext(ctx context.Context) IConvTurnDo {
	return c.withDO(c.DO.WithContext(ctx))
}

func (c *convTurnDo) withDO(do gen.Dao) *convTurnDo {
	c.DO = *do.(*gen.DO)
	return c
}