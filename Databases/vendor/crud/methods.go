package crud

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// GetTables returns list of table names
func (db *Agent) GetTables() document {
	var tables []string
	for tb := range db.Schema {
		tables = append(tables, tb)
	}
	sort.Strings(tables)

	return document{
		"tables": tables,
	}
}

// GetRows reads from table with constraints
func (db *Agent) GetRows(
	table, limit, offset string,
) (document, error) {
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
		result  = []document{}
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
			doc = make(document, len(columns))
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

	return document{
		"records": result,
	}, nil
}

// GetRow returns row using <id> from table
func (db *Agent) GetRow(
	table, id string,
) (document, error) {
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
		doc = make(document, len(columns))
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

	return document{
		"record": doc,
	}, nil
}

// NewRow adds entity to existing table
func (db *Agent) NewRow(
	table string, data document,
) (document, error) {
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
		return nil, err
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	d := document{}
	d[getPrimaryKey(db.Schema, table)] = insertID
	return d, nil
}

// UpdateRow updates row using <id> and valid r.body <doc>
func (db *Agent) UpdateRow(
	table, id string, data document,
) (document, error) {
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
		return nil, err
	}

	updateID, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	return document{
		"updated": updateID,
	}, nil
}

// DeleteRow deletes row using primary key
func (db *Agent) DeleteRow(
	table, id string,
) (document, error) {
	q := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ?",
		table, getPrimaryKey(db.Schema, table),
	)
	res, err := db.Exec(q, id)
	if err != nil {
		return nil, err
	}
	deleteID, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	return document{
		"deleted": deleteID,
	}, nil
}
