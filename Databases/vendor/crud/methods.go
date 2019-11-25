package crud

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// GetTables returns tables name list
func (db *Agent) GetTables() []string {
	var tables []string
	for tb := range db.Schema {
		tables = append(tables, tb)
	}
	sort.Strings(tables)

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

			doc[columns[i].name] = val
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

		doc[columns[i].name] = val
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
