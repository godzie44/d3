package schema

import (
	"database/sql"
	"fmt"
	"github.com/godzie44/d3/orm/entity"
	"reflect"
	"strings"
	"time"
)

type ColumnType string

const (
	Bool  ColumnType = "bool"
	Int   ColumnType = "int"
	Int32 ColumnType = "int32"
	Int64 ColumnType = "int64"

	Float64 ColumnType = "float64"
	Float32 ColumnType = "float32"

	String ColumnType = "string"

	Time ColumnType = "time"

	UUID ColumnType = "uuid"

	NullBool  ColumnType = "null-bool"
	NullInt32 ColumnType = "null-int32"
	NullInt64 ColumnType = "null-int64"

	NullFloat64 ColumnType = "null-float64"

	NullString ColumnType = "null-string"

	NullTime ColumnType = "null-time"
)

func toNullEquivalent(t ColumnType) ColumnType {
	switch t {
	case Bool:
		return NullBool
	case Int:
		return NullInt64
	case Int32:
		return NullInt32
	case Int64:
		return NullInt64
	case Float32:
		return NullFloat64
	case Float64:
		return NullFloat64
	case String:
		return NullString
	case Time:
		return NullTime
	}
	return t
}

func toNotNullEquivalent(t ColumnType) ColumnType {
	switch t {
	case NullBool:
		return Bool
	case NullInt32:
		return Int32
	case NullInt64:
		return Int64
	case NullFloat64:
		return Float64
	case NullString:
		return String
	case NullTime:
		return Time
	}
	return t
}

type StorageSchemaGenerator interface {
	CreateTableSql(name string, columns map[string]ColumnType, pkColumns []string, pkStrategy entity.PkStrategy) string
	CreateIndexSql(name string, unique bool, table string, columns ...string) string
}

type Builder struct {
	schemaBuilder StorageSchemaGenerator
}

func NewBuilder(schemaBuilder StorageSchemaGenerator) *Builder {
	return &Builder{schemaBuilder: schemaBuilder}
}

func (b *Builder) Build(registry *entity.MetaRegistry) (string, error) {
	createTableCommands, err := b.createNewTableCommands(registry)
	if err != nil {
		return "", err
	}

	var res strings.Builder
	for _, cmd := range createTableCommands {
		res.WriteString(b.schemaBuilder.CreateTableSql(cmd.tableName, cmd.columns, cmd.pkColumns, cmd.pkStrategy))

		for _, ind := range cmd.indexes {
			res.WriteString(b.schemaBuilder.CreateIndexSql(ind.Name, ind.Unique, cmd.tableName, ind.Columns...))
		}
	}

	return res.String(), nil
}

type newTableCmd struct {
	tableName  string
	columns    map[string]ColumnType
	pkColumns  []string
	pkStrategy entity.PkStrategy
	indexes    []entity.Index
}

func (b *Builder) createNewTableCommands(registry *entity.MetaRegistry) (map[entity.Name]*newTableCmd, error) {
	createTableCmdQueue := make(map[entity.Name]*newTableCmd)

	err := registry.ForEach(func(meta *entity.MetaInfo) error {
		if _, exists := createTableCmdQueue[meta.EntityName]; !exists {
			createTableCmdQueue[meta.EntityName] = &newTableCmd{
				columns: make(map[string]ColumnType),
			}
		}

		createTableCommand := createTableCmdQueue[meta.EntityName]
		createTableCommand.indexes = meta.Indexes
		createTableCommand.tableName = meta.TableName

		for _, field := range meta.Fields {
			colType, err := reflectTypeToDbType(field.AssociatedType)
			if err != nil {
				return err
			}

			createTableCommand.columns[field.DbAlias] = colType
		}

		createTableCommand.pkColumns = []string{meta.Pk.Field.DbAlias}
		createTableCommand.pkStrategy = meta.Pk.Strategy

		for _, rel := range meta.OneToOneRelations() {
			relatedMeta := meta.RelatedMeta[rel.RelatedWith()]
			colType, err := reflectTypeToDbType(relatedMeta.Pk.Field.AssociatedType)
			if err != nil {
				return err
			}

			switch rel.DeleteStrategy() {
			case entity.None, entity.Cascade:
				createTableCommand.columns[rel.JoinColumn] = colType
			case entity.Nullable:
				createTableCommand.columns[rel.JoinColumn] = toNullEquivalent(colType)
			}
		}

		for _, rel := range meta.OneToManyRelations() {
			if _, exists := createTableCmdQueue[rel.RelatedWith()]; !exists {
				createTableCmdQueue[rel.RelatedWith()] = &newTableCmd{
					columns: make(map[string]ColumnType),
				}
			}

			colType, err := reflectTypeToDbType(meta.Pk.Field.AssociatedType)
			if err != nil {
				return err
			}

			switch rel.DeleteStrategy() {
			case entity.None, entity.Cascade:
				createTableCmdQueue[rel.RelatedWith()].columns[rel.JoinColumn] = colType
			case entity.Nullable:
				createTableCmdQueue[rel.RelatedWith()].columns[rel.JoinColumn] = toNullEquivalent(colType)
			}
		}

		for _, rel := range meta.ManyToManyRelations() {
			if _, exists := createTableCmdQueue[entity.Name(rel.JoinTable)]; exists {
				continue
			}

			joinColType, err := reflectTypeToDbType(meta.Pk.Field.AssociatedType)
			if err != nil {
				return err
			}

			refColType, err := reflectTypeToDbType(meta.RelatedMeta[rel.RelatedWith()].Pk.Field.AssociatedType)
			if err != nil {
				return err
			}

			createTableCmdQueue[entity.Name(rel.JoinTable)] = &newTableCmd{
				tableName: rel.JoinTable,
				columns:   map[string]ColumnType{rel.JoinColumn: toNotNullEquivalent(joinColType), rel.ReferenceColumn: toNotNullEquivalent(refColType)},
				pkColumns: []string{rel.JoinColumn, rel.ReferenceColumn},
			}
		}

		return nil
	})

	return createTableCmdQueue, err
}

func reflectTypeToDbType(t reflect.Type) (ColumnType, error) {
	if t.Name() == "UUID" {
		return UUID, nil
	}

	switch t.Kind() {
	case reflect.Bool:
		return Bool, nil
	case reflect.Int:
		return Int, nil
	case reflect.Int32:
		return Int32, nil
	case reflect.Int64:
		return Int64, nil
	case reflect.Float32:
		return Float32, nil
	case reflect.Float64:
		return Float64, nil
	case reflect.String:
		return String, nil
	}

	if t.AssignableTo(reflect.TypeOf(time.Time{})) {
		return Time, nil
	}

	if t.Kind() == reflect.Struct {
		switch {
		case t.AssignableTo(reflect.TypeOf(sql.NullBool{})):
			return NullBool, nil
		case t.AssignableTo(reflect.TypeOf(sql.NullInt32{})):
			return NullInt32, nil
		case t.AssignableTo(reflect.TypeOf(sql.NullInt64{})):
			return NullInt64, nil
		case t.AssignableTo(reflect.TypeOf(sql.NullFloat64{})):
			return NullFloat64, nil
		case t.AssignableTo(reflect.TypeOf(sql.NullString{})):
			return NullString, nil
		case t.AssignableTo(reflect.TypeOf(sql.NullTime{})):
			return NullTime, nil
		}
	}

	return Int, fmt.Errorf("unsupported field type")
}
