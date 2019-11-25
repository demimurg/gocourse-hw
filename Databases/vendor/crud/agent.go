package crud

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Column has meta of DB column
type Column struct {
	Name     string
	Type     reflect.Type
	Nullable bool
}

// Table represents db-table
type Table []Column

// DbSchema describes database structure for app
type DbSchema map[string]Table

// Agent covers CRUD ops on db
type Agent struct {
	*sql.DB
	Schema DbSchema
}

type validationErr struct {
	Field string
}

func (e validationErr) Error() string {
	// should be "field <f> have invalid type"
	return fmt.Sprintf(
		"field %s have invalid type",
		e.Field,
	)
}

// ReadDbSchema saves tables/columns meta to the receiver
// DbAgent Initialization
func (db *Agent) ReadDbSchema() error {
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

// Validate ...
func (db *Agent) Validate(
	table, method string, data map[string]interface{},
) error {
	// waiting for refactor...
	// !!!HARDCODE!!!
	prKey := getPrimaryKey(db.Schema, table)
	_, havePrimaryKey := data[prKey]
	if havePrimaryKey && method == "UPDATE" {
		return validationErr{prKey}
	}
	// !!!HARDCODE!!!

	for _, col := range db.Schema[table][1:] {
		_, inReq := data[col.Name]
		if !inReq && method == "UPDATE" {
			continue
		}

		if !inReq && !col.Nullable {
			switch col.Type.Name() {
			case "int32":
				data[col.Name] = 0
			case "RawBytes":
				data[col.Name] = ""
			}
		}

		var (
			val       = data[col.Name]
			nullValue = col.Nullable && val == nil
			err       error
		)
		switch col.Type.Name() {
		case "int32":
			if _, ok := val.(int); !nullValue && !ok {
				err = validationErr{col.Name}
			}
		case "RawBytes":
			if _, ok := val.(string); !nullValue && !ok {
				err = validationErr{col.Name}
			}
		default:
			err = validationErr{col.Name}
		}

		if err != nil {
			return err
		}
	}

	for k := range data {
		var inSchema bool
		for _, col := range db.Schema[table][1:] {
			if k == col.Name {
				inSchema = true
				break
			}
		}

		if !inSchema {
			delete(data, k)
		}
	}

	return nil
}

func getPrimaryKey(schema DbSchema, table string) string {
	var key string
	for _, column := range schema[table] {
		if strings.HasSuffix(column.Name, "id") {
			key = column.Name
			break
		}
	}
	return key
}

// GetTables returns tables name list
func (db *Agent) GetTables() []string {
	var tables []string
	for tb := range db.Schema {
		tables = append(tables, tb)
	}

	return tables
}

// GetRows ...
func (db *Agent) GetRows(
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

// GetRow returns row <id> from table
func (db *Agent) GetRow(table, id string) (
	map[string]interface{}, error,
) {
	q := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s = ?",
		table, getPrimaryKey(db.Schema, table),
	)
	row := db.QueryRow(q, id)

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
func (db *Agent) NewRow(
	table string, data map[string]interface{},
) (string, int64, error) {
	err := db.Validate(table, "CREATE", data)
	if err != nil {
		return "", 0, err
	}

	var (
		columns = make([]string, len(data))
		values  = make([]interface{}, len(data))
		i       int
	)
	for k, v := range data {
		columns[i] = k
		values[i] = fmt.Sprintf("%v", v)
		i++
	}

	queryDraft := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "),
		"?"+strings.Repeat(", ?", len(columns)-1), // variadic num of placeholders
	)

	res, err := db.Exec(queryDraft, values...)
	if err != nil {
		fmt.Println(err)
		return "", 0, err
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return "", 0, err
	}

	return getPrimaryKey(db.Schema, table), insertID, nil
}

// UpdateRow ...
func (db *Agent) UpdateRow(
	table, id string, data map[string]interface{},
) (int64, error) {
	err := db.Validate(table, "UPDATE", data)
	if err != nil {
		return 0, err
	}

	var (
		valsWithID = make([]interface{}, len(data)+1)
		pairs      string
		i          int
	)
	for k, v := range data {
		pairs += fmt.Sprintf("%s = ?, ", k)
		valsWithID[i] = v
		i++
	}
	valsWithID[i] = id
	pairs = strings.TrimRight(pairs, ", ")

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ?",
		table, pairs,
		getPrimaryKey(db.Schema, table),
	)
	res, err := db.Exec(query, valsWithID...)
	if err != nil {
		return 0, err
	}

	updateID, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return updateID, nil
}

// DeleteRow ...
func (db *Agent) DeleteRow(table, id string) (int64, error) {
	q := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		table, getPrimaryKey(db.Schema, table),
	)
	res, err := db.Exec(q, id)
	if err != nil {
		return 0, err
	}
	deleteID, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deleteID, nil
}
