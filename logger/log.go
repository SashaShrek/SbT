package logger

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

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

func Take(typ string, fields map[string]string, text string) {
	logs := log.Fields{}
	for key, value := range fields {
		logs[key] = value
	}
	var file, err = os.OpenFile("logs/"+typ+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Could Not Open Log File : " + err.Error())
	}
	log.SetOutput(file)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	switch typ {
	case "info":
		log.WithFields(logs).Info(text)
	case "warn":
		log.WithFields(logs).Warn(text)
	case "error":
		log.WithFields(logs).Error(text)
	}
}
