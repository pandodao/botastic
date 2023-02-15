// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package dao

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gen/helper"

	"gorm.io/plugin/dbresolver"

	"github.com/pandodao/botastic/core"
)

func newProperty(db *gorm.DB, opts ...gen.DOOption) property {
	_property := property{}

	_property.propertyDo.UseDB(db, opts...)
	_property.propertyDo.UseModel(&core.Property{})

	tableName := _property.propertyDo.TableName()
	_property.ALL = field.NewAsterisk(tableName)
	_property.Key = field.NewString(tableName, "key")
	_property.Value = field.NewString(tableName, "value")
	_property.UpdatedAt = field.NewTime(tableName, "updated_at")

	_property.fillFieldMap()

	return _property
}

type property struct {
	propertyDo

	ALL       field.Asterisk
	Key       field.String
	Value     field.String
	UpdatedAt field.Time

	fieldMap map[string]field.Expr
}

func (p property) Table(newTableName string) *property {
	p.propertyDo.UseTable(newTableName)
	return p.updateTableName(newTableName)
}

func (p property) As(alias string) *property {
	p.propertyDo.DO = *(p.propertyDo.As(alias).(*gen.DO))
	return p.updateTableName(alias)
}

func (p *property) updateTableName(table string) *property {
	p.ALL = field.NewAsterisk(table)
	p.Key = field.NewString(table, "key")
	p.Value = field.NewString(table, "value")
	p.UpdatedAt = field.NewTime(table, "updated_at")

	p.fillFieldMap()

	return p
}

func (p *property) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := p.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (p *property) fillFieldMap() {
	p.fieldMap = make(map[string]field.Expr, 3)
	p.fieldMap["key"] = p.Key
	p.fieldMap["value"] = p.Value
	p.fieldMap["updated_at"] = p.UpdatedAt
}

func (p property) clone(db *gorm.DB) property {
	p.propertyDo.ReplaceConnPool(db.Statement.ConnPool)
	return p
}

func (p property) replaceDB(db *gorm.DB) property {
	p.propertyDo.ReplaceDB(db)
	return p
}

type propertyDo struct{ gen.DO }

type IPropertyDo interface {
	gen.SubQuery
	Debug() IPropertyDo
	WithContext(ctx context.Context) IPropertyDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IPropertyDo
	WriteDB() IPropertyDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IPropertyDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IPropertyDo
	Not(conds ...gen.Condition) IPropertyDo
	Or(conds ...gen.Condition) IPropertyDo
	Select(conds ...field.Expr) IPropertyDo
	Where(conds ...gen.Condition) IPropertyDo
	Order(conds ...field.Expr) IPropertyDo
	Distinct(cols ...field.Expr) IPropertyDo
	Omit(cols ...field.Expr) IPropertyDo
	Join(table schema.Tabler, on ...field.Expr) IPropertyDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IPropertyDo
	RightJoin(table schema.Tabler, on ...field.Expr) IPropertyDo
	Group(cols ...field.Expr) IPropertyDo
	Having(conds ...gen.Condition) IPropertyDo
	Limit(limit int) IPropertyDo
	Offset(offset int) IPropertyDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IPropertyDo
	Unscoped() IPropertyDo
	Create(values ...*core.Property) error
	CreateInBatches(values []*core.Property, batchSize int) error
	Save(values ...*core.Property) error
	First() (*core.Property, error)
	Take() (*core.Property, error)
	Last() (*core.Property, error)
	Find() ([]*core.Property, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*core.Property, err error)
	FindInBatches(result *[]*core.Property, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*core.Property) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IPropertyDo
	Assign(attrs ...field.AssignExpr) IPropertyDo
	Joins(fields ...field.RelationField) IPropertyDo
	Preload(fields ...field.RelationField) IPropertyDo
	FirstOrInit() (*core.Property, error)
	FirstOrCreate() (*core.Property, error)
	FindByPage(offset int, limit int) (result []*core.Property, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IPropertyDo
	UnderlyingDB() *gorm.DB
	schema.Tabler

	Get(ctx context.Context, key string) (result string, err error)
	Set(ctx context.Context, key string, value interface{}) (result int64, err error)
}

// SELECT value FROM @@table WHERE key=@key
func (p propertyDo) Get(ctx context.Context, key string) (result string, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	params = append(params, key)
	generateSQL.WriteString("SELECT value FROM properties WHERE key=? ")

	var executeSQL *gorm.DB
	executeSQL = p.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

// UPDATE @@table
// {{set}}
//
//	value=@value,
//	updated_at=NOW()
//
// {{end}}
// WHERE key=@key
func (p propertyDo) Set(ctx context.Context, key string, value interface{}) (result int64, err error) {
	var params []interface{}

	var generateSQL strings.Builder
	generateSQL.WriteString("UPDATE properties ")
	var setSQL0 strings.Builder
	params = append(params, value)
	setSQL0.WriteString("value=?, updated_at=NOW() ")
	helper.JoinSetBuilder(&generateSQL, setSQL0)
	params = append(params, key)
	generateSQL.WriteString("WHERE key=? ")

	var executeSQL *gorm.DB
	executeSQL = p.UnderlyingDB().Raw(generateSQL.String(), params...).Take(&result) // ignore_security_alert
	err = executeSQL.Error

	return
}

func (p propertyDo) Debug() IPropertyDo {
	return p.withDO(p.DO.Debug())
}

func (p propertyDo) WithContext(ctx context.Context) IPropertyDo {
	return p.withDO(p.DO.WithContext(ctx))
}

func (p propertyDo) ReadDB() IPropertyDo {
	return p.Clauses(dbresolver.Read)
}

func (p propertyDo) WriteDB() IPropertyDo {
	return p.Clauses(dbresolver.Write)
}

func (p propertyDo) Session(config *gorm.Session) IPropertyDo {
	return p.withDO(p.DO.Session(config))
}

func (p propertyDo) Clauses(conds ...clause.Expression) IPropertyDo {
	return p.withDO(p.DO.Clauses(conds...))
}

func (p propertyDo) Returning(value interface{}, columns ...string) IPropertyDo {
	return p.withDO(p.DO.Returning(value, columns...))
}

func (p propertyDo) Not(conds ...gen.Condition) IPropertyDo {
	return p.withDO(p.DO.Not(conds...))
}

func (p propertyDo) Or(conds ...gen.Condition) IPropertyDo {
	return p.withDO(p.DO.Or(conds...))
}

func (p propertyDo) Select(conds ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Select(conds...))
}

func (p propertyDo) Where(conds ...gen.Condition) IPropertyDo {
	return p.withDO(p.DO.Where(conds...))
}

func (p propertyDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) IPropertyDo {
	return p.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (p propertyDo) Order(conds ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Order(conds...))
}

func (p propertyDo) Distinct(cols ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Distinct(cols...))
}

func (p propertyDo) Omit(cols ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Omit(cols...))
}

func (p propertyDo) Join(table schema.Tabler, on ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Join(table, on...))
}

func (p propertyDo) LeftJoin(table schema.Tabler, on ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.LeftJoin(table, on...))
}

func (p propertyDo) RightJoin(table schema.Tabler, on ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.RightJoin(table, on...))
}

func (p propertyDo) Group(cols ...field.Expr) IPropertyDo {
	return p.withDO(p.DO.Group(cols...))
}

func (p propertyDo) Having(conds ...gen.Condition) IPropertyDo {
	return p.withDO(p.DO.Having(conds...))
}

func (p propertyDo) Limit(limit int) IPropertyDo {
	return p.withDO(p.DO.Limit(limit))
}

func (p propertyDo) Offset(offset int) IPropertyDo {
	return p.withDO(p.DO.Offset(offset))
}

func (p propertyDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IPropertyDo {
	return p.withDO(p.DO.Scopes(funcs...))
}

func (p propertyDo) Unscoped() IPropertyDo {
	return p.withDO(p.DO.Unscoped())
}

func (p propertyDo) Create(values ...*core.Property) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Create(values)
}

func (p propertyDo) CreateInBatches(values []*core.Property, batchSize int) error {
	return p.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (p propertyDo) Save(values ...*core.Property) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Save(values)
}

func (p propertyDo) First() (*core.Property, error) {
	if result, err := p.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*core.Property), nil
	}
}

func (p propertyDo) Take() (*core.Property, error) {
	if result, err := p.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*core.Property), nil
	}
}

func (p propertyDo) Last() (*core.Property, error) {
	if result, err := p.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*core.Property), nil
	}
}

func (p propertyDo) Find() ([]*core.Property, error) {
	result, err := p.DO.Find()
	return result.([]*core.Property), err
}

func (p propertyDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*core.Property, err error) {
	buf := make([]*core.Property, 0, batchSize)
	err = p.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (p propertyDo) FindInBatches(result *[]*core.Property, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return p.DO.FindInBatches(result, batchSize, fc)
}

func (p propertyDo) Attrs(attrs ...field.AssignExpr) IPropertyDo {
	return p.withDO(p.DO.Attrs(attrs...))
}

func (p propertyDo) Assign(attrs ...field.AssignExpr) IPropertyDo {
	return p.withDO(p.DO.Assign(attrs...))
}

func (p propertyDo) Joins(fields ...field.RelationField) IPropertyDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Joins(_f))
	}
	return &p
}

func (p propertyDo) Preload(fields ...field.RelationField) IPropertyDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Preload(_f))
	}
	return &p
}

func (p propertyDo) FirstOrInit() (*core.Property, error) {
	if result, err := p.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*core.Property), nil
	}
}

func (p propertyDo) FirstOrCreate() (*core.Property, error) {
	if result, err := p.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*core.Property), nil
	}
}

func (p propertyDo) FindByPage(offset int, limit int) (result []*core.Property, count int64, err error) {
	result, err = p.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = p.Offset(-1).Limit(-1).Count()
	return
}

func (p propertyDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = p.Count()
	if err != nil {
		return
	}

	err = p.Offset(offset).Limit(limit).Scan(result)
	return
}

func (p propertyDo) Scan(result interface{}) (err error) {
	return p.DO.Scan(result)
}

func (p propertyDo) Delete(models ...*core.Property) (result gen.ResultInfo, err error) {
	return p.DO.Delete(models)
}

func (p *propertyDo) withDO(do gen.Dao) *propertyDo {
	p.DO = *do.(*gen.DO)
	return p
}
