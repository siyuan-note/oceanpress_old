package sqlite

import (
	"database/sql"

	"github.com/2234839/md2website/src/util"
	_ "github.com/mattn/go-sqlite3"
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
		if err != nil {
			util.Warn(err)
			return ids
		}
		defer rows.Close()

		columns, _ := rows.Columns()
		columnLength := len(columns)
		cache := make([]interface{}, columnLength) //临时存储每行数据
		for index, _ := range cache {              //为每一列初始化一个指针
			var a interface{}
			cache[index] = &a
		}
		var list []map[string]interface{} //返回的切片
		for rows.Next() {
			err := rows.Scan(cache...)

			item := make(map[string]interface{})
			for i, data := range cache {
				item[columns[i]] = *data.(*interface{}) //取实际类型
			}
			list = append(list, item)
			if err != nil {
				util.Warn(err)
			} else {
				ids = append(ids, item["id"].(string))
			}
		}
		// util.Warn(list)
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
