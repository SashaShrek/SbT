package logger

import (
	"fmt"

	db "github.com/SashaShrek/db"
)

func SetLog(user_id string, type_log string, service string, text string) {
	var query string
	if user_id == "-1" {
		query = fmt.Sprintf("insert into logs (row_id, user_id, date, type, service, text)"+
			"values ((select max(row_id) + 1 from logs), -1, current_timestamp, '%s', '%s',"+
			"'%s')", type_log, service, text)
	} else {
		query = fmt.Sprintf("insert into logs (row_id, user_id, date, type, service, text)"+
			"values ((select max(row_id) + 1 from logs), (select row_id from users where tlgrm_id = '%s'), current_timestamp, '%s', '%s',"+
			"'%s')", user_id, type_log, service, text)
	}
	_ = db.InsertOrUpdate(query)
}
