package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/siyuan-note/oceanpress/src/util"
)

// DbResult 初始化数据库的结果
type DbResult struct {
	SQLToID func(string) []string
}

// InitDb 初始化数据库
func InitDb(dbPath string) DbResult {
	db, err := sql.Open("sqlite3", dbPath)

	checkErr(err)

	sqlToID := func(sql string) []string {
		var ids []string

		rows, err := db.Query(sql)
		cols, _ := rows.Columns()
		defer rows.Close()
		if err != nil {
			util.Warn("sql 查询错误", err)
			return ids
		}
		for rows.Next() {
			// Create a slice of interface{}'s to represent each column,
			// and a second slice to contain pointers to each item in the columns slice.
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i, _ := range columns {
				columnPointers[i] = &columns[i]
			}

			// Scan the result into the column pointers...
			if err := rows.Scan(columnPointers...); err != nil {
				util.Warn("sql 结果 scan 错误", err)
			} else {
				// Create our map, and retrieve the value for each column from the pointers slice,
				// storing it in the map with the name of the column as the key.
				m := make(map[string]interface{})
				for i, colName := range cols {
					val := columnPointers[i].(*interface{})
					m[colName] = *val
				}
				id := m["id"].(string)
				if len(id) > 0 {
					ids = append(ids, id)
				}
			}
		}
		return ids
	}

	return DbResult{
		SQLToID: sqlToID,
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
