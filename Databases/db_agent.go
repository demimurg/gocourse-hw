package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Column have meta of DB column
type Column struct {
	Name     string
	Type     reflect.Type
	Nullable bool
}

// Table represent db-table
type Table []Column

// DbSchema describes database structure for app
type DbSchema map[string]Table

// DbAgent covers CRUD ops on db
type DbAgent struct {
	*sql.DB
	Schema DbSchema
}

type validationErr struct {
	Value, Column, Type string
}

func (e validationErr) Error() string {
	// should be "field <f> have invalid type"
	return fmt.Sprintf(
		"<%s> wrong value for <%s> column, must be %s",
		e.Value, e.Column, e.Type,
	)
}

// ReadDbSchema save tables/columns meta to the receiver
// DbAgent Initialization
func (db *DbAgent) ReadDbSchema() error {
	schema := DbSchema{}

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return err
	}

	var tbName string
	for rows.Next() {
		rows.Scan(&tbName)

		schema[tbName] = Table{}
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
			schema[table] = append(schema[table], Column{
				meta.Name(), meta.ScanType(), isNullable,
			})
		}
		rows.Close()
	}

	db.Schema = schema
	return nil
}

// ValidForm returns valid map with data from url.Values
// Uses columns meta to parse data
func (db *DbAgent) ValidForm(
	table string,
	vals url.Values,
) (map[string]string, error) {
	// Создание новой мапки добавлят нагрузку. Стоит ли?
	columns, _ := db.Schema[table]

	form := make(map[string]string, len(vals))
	for _, col := range columns {
		v := vals.Get(col.Name)
		if v == "" && !col.Nullable {
			return nil, validationErr{v, col.Name, "not null"}
		}

		switch col.Type.Name() {
		case "int":
			if _, e := strconv.Atoi(v); e != nil {
				return nil, validationErr{v, col.Name, "int"}
			}
		case "float64":
			if _, e := strconv.ParseFloat(v, 64); e != nil {
				return nil, validationErr{v, col.Name, "float64"}
			}
		}

		form[col.Name] = v
	}
	return form, nil
}

// GetTables return tables name list
func (db *DbAgent) GetTables() []string {
	var tables []string
	for tb := range db.Schema {
		tables = append(tables, tb)
	}

	return tables
}

// GetRows ...
func (db *DbAgent) GetRows(
	table, limit, offset string,
) ([]map[string]interface{}, error) {
	if l, err := strconv.Atoi(limit); err != nil || l <= 0 {
		limit = "5"
	}
	if o, err := strconv.Atoi(offset); err != nil || o < 0 {
		offset = "0"
	}

	rows, err := db.Query(
		"SELECT * FROM "+table+
			" LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		result  = []map[string]interface{}{}
		columns = db.Schema[table]
	)
	for rows.Next() {
		vals := make([]interface{}, len(columns))
		valPtrs := make([]interface{}, len(columns))
		for i := range columns {
			valPtrs[i] = &vals[i]
		}

		err = rows.Scan(valPtrs...)
		if err != nil {
			return nil, err
		}

		var (
			doc = make(map[string]interface{}, len(columns))
			val interface{}
		)
		for i := range vals {
			switch v := vals[i].(type) {
			case []byte:
				val = string(v)
			case int64:
				val = v
			default:
				val = nil
			}

			doc[columns[i].Name] = val
		}
		result = append(result, doc)
	}

	return result, nil
}

// GetRow return row <id> from table
func (db *DbAgent) GetRow(table, id string) (
	map[string]interface{}, error,
) {
	row := db.QueryRow(
		"SELECT * FROM "+table+
			" WHERE id = ?", id,
	)

	var (
		columns = db.Schema[table]
		vals    = make([]interface{}, len(columns))
		valPtrs = make([]interface{}, len(columns))
	)
	for i := range columns {
		valPtrs[i] = &vals[i]
	}
	err := row.Scan(valPtrs...)
	if err != nil {
		return nil, err
	}

	var (
		doc = make(map[string]interface{}, len(columns))
		val interface{}
	)

	for i := range vals {
		switch v := vals[i].(type) {
		case []byte:
			val = string(v)
		case int64:
			val = v
		default:
			val = nil
		}

		doc[columns[i].Name] = val
	}

	return doc, nil
}

// NewRow adds entity to existing table
func (db *DbAgent) NewRow(table string, vals url.Values) (int64, error) {
	// обрабатывать неверную таблицу (по уму)
	form, err := db.ValidForm(table, vals)
	if err != nil {
		return 0, err
	}

	fields := make([]string, len(form))
	values := make([]string, len(form))
	for k, v := range form {
		fields = append(fields, k)
		values = append(values, v)
	}

	res, err := db.Exec(
		"INSERT INTO "+table+" (?) VALUES (?)",
		strings.Join(fields, ", "),
		strings.Join(values, ", "),
	)

	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return insertID, nil
}

// UpdateRow ...
func (db *DbAgent) UpdateRow(
	table, id string, vals url.Values,
) (int64, error) {
	form, err := db.ValidForm(table, vals)
	if err != nil {
		return 0, err
	}

	var subcmd string
	for k, v := range form {
		subcmd += fmt.Sprintf("%s='%s', ", k, v)
	}
	subcmd = subcmd[:len(subcmd)-2]

	res, err := db.Exec(
		"UPDATE "+table+" SET ? WHERE id = ?",
		subcmd, id,
	)

	updateID, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return updateID, nil
}

// DeleteRow ...
func (db *DbAgent) DeleteRow(table, id string) (int64, error) {
	res, err := db.Exec("DELETE FROM "+table+" WHERE id = ?", id)
	if err != nil {
		return 0, err
	}
	deleteID, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deleteID, nil
}
