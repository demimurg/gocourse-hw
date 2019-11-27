package crud

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// Agent covers CRUD ops on db
type Agent struct {
	*sql.DB
	// schema should be created only with ReadDbSchema method, but
	// you can see if table in schema <_, ok := a.Schema>
	Schema dbSchema
}

type dbSchema map[string]table

type table []column

type column struct {
	name     string
	typ      reflect.Type
	nullable bool
}

type validationErr struct {
	Field string
}

func (e validationErr) Error() string {
	return fmt.Sprintf(
		"field %s have invalid type",
		e.Field,
	)
}

type document map[string]interface{}

// Validate checks r.body from PUT and POST
// returns error if primary key in "update" body
// uses type <document> to be sure that crud methods have valid data
func (db *Agent) Validate(
	table, method string, data map[string]interface{},
) (document, error) {
	prKey := getPrimaryKey(db.Schema, table)
	_, havePrimaryKey := data[prKey]
	if havePrimaryKey && method == "UPDATE" {
		return nil, validationErr{prKey}
	}

	for _, col := range db.Schema[table][1:] {
		_, inReq := data[col.name]
		if !inReq && method == "UPDATE" {
			continue
		}

		if !inReq && !col.nullable {
			switch col.typ.Name() {
			case "int32":
				data[col.name] = 0
			case "RawBytes":
				data[col.name] = ""
			}
		}

		var (
			val       = data[col.name]
			nullValue = col.nullable && val == nil
			err       error
		)
		switch col.typ.Name() {
		case "int32":
			if _, ok := val.(int); !nullValue && !ok {
				err = validationErr{col.name}
			}
		case "RawBytes":
			if _, ok := val.(string); !nullValue && !ok {
				err = validationErr{col.name}
			}
		default:
			err = validationErr{col.name}
		}

		if err != nil {
			return nil, err
		}
	}

	for k := range data {
		var inSchema bool
		for _, col := range db.Schema[table][1:] {
			if k == col.name {
				inSchema = true
				break
			}
		}

		if !inSchema {
			delete(data, k)
		}
	}

	return document(data), nil
}

// ReadDbSchema saves tables/columns meta to the receiver
// DbAgent Initialization
func (db *Agent) ReadDbSchema() error {
	schema := dbSchema{}

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return err
	}

	var tbName string
	for rows.Next() {
		rows.Scan(&tbName)

		schema[tbName] = table{}
	}
	rows.Close()

	for table := range schema {
		rows, err := db.Query("SELECT * FROM " + table)
		if err != nil {
			return err
		}

		cols, err := rows.ColumnTypes()
		if err != nil {
			return err
		}

		for _, meta := range cols {
			isNullable, _ := meta.Nullable()
			schema[table] = append(schema[table], column{
				meta.Name(), meta.ScanType(), isNullable,
			})
		}
		rows.Close()
	}

	db.Schema = schema
	return nil
}

func getPrimaryKey(schema dbSchema, table string) string {
	var key string
	for _, column := range schema[table] {
		if strings.HasSuffix(column.name, "id") {
			key = column.name
			break
		}
	}
	return key
}
