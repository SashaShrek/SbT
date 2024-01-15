package cusers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sbt/logger"
	"strconv"
	"time"

	"github.com/SashaShrek/db"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func Appr(bot *tgbotapi.BotAPI, token *string, idChannel *int64) {
	type appInv struct {
		tlgrm_id string
	}
	type DataAprrove struct {
		result  bool
		desc    string
		id      int64
		aInv    appInv
		err     error
		data    []byte
		res     *http.Response
		message tgbotapi.MessageConfig
	}

	var app DataAprrove

	for range time.Tick(15 * time.Minute) {
		rows, datab, _ := db.Select("select tlgrm_id from users where invite = true")
		for rows.Next() {
			app.err = rows.Scan(&app.aInv.tlgrm_id)
			if app.err != nil {
				continue
			}
			app.id, _ = strconv.ParseInt(app.aInv.tlgrm_id, 10, 64)
			app.res, app.err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/approveChatJoinRequest?chat_id=%d&user_id=%d", *token, *idChannel, app.id))
			if app.err != nil {
				log := map[string]string{
					"User": app.aInv.tlgrm_id,
					"Func": "Appr",
				}
				logger.Take("error", log, app.err.Error())
				logger.SetLog(app.aInv.tlgrm_id, "error", "approve", app.err.Error())
				continue
			}
			var dataRes map[string]interface{}
			app.data, _ = ioutil.ReadAll(app.res.Body)
			app.err = json.Unmarshal(app.data, &dataRes)
			if app.err != nil {
				log := map[string]string{
					"User": app.aInv.tlgrm_id,
					"Func": "Appr",
				}
				logger.Take("error", log, app.err.Error())
			}

			app.result = dataRes["ok"].(bool)
			if dataRes["description"] != nil {
				app.desc = dataRes["description"].(string)
			}
			if !app.result {
				if app.desc == "Bad Request: USER_ALREADY_PARTICIPANT" {
					app.err = db.InsertOrUpdate(fmt.Sprintf("update users set invite = false where tlgrm_id = '%s'", app.aInv.tlgrm_id))
					if app.err != nil {
						log := map[string]string{
							"User": app.aInv.tlgrm_id,
							"Func": "Appr",
						}
						logger.Take("error", log, app.err.Error())
						logger.SetLog(app.aInv.tlgrm_id, "error", "approve", app.err.Error())
						app.res.Body.Close()
					}
				} else {
					log := map[string]string{
						"User": app.aInv.tlgrm_id,
						"Func": "Appr",
					}
					logger.Take("warn", log, app.desc)
					logger.SetLog(app.aInv.tlgrm_id, "warn", "approve", app.desc)
				}
				/*message := tgbotapi.NewMessage(id, "Возникла ошибка при одобрении вашей заявки. Вероятно, у вас ссылка старого типа. Напишите на почту supp.sbt@gmail.com с просьбой обновить ссылку.")
				bot.Send(message)*/
				/*app.err = db.InsertOrUpdate(fmt.Sprintf("update users set invite = false where tlgrm_id = '%s'", app.aInv.tlgrm_id))
				if app.err != nil {
					logger.SetLog(app.aInv.tlgrm_id, "error", "approve", app.err.Error())
				}*/
				app.res.Body.Close()
				continue
			}
			app.err = db.InsertOrUpdate(fmt.Sprintf("update users set invite = false where tlgrm_id = '%s'", app.aInv.tlgrm_id))
			if app.err != nil {
				log := map[string]string{
					"User": app.aInv.tlgrm_id,
					"Func": "Appr",
				}
				logger.Take("error", log, app.err.Error())
				logger.SetLog(app.aInv.tlgrm_id, "error", "approve", app.err.Error())
				app.res.Body.Close()
				continue
			}
			app.message = tgbotapi.NewMessage(app.id, "Доступ к каналу предоставлен: заявка одобрена")
			bot.Send(app.message)
			logger.SetLog(app.aInv.tlgrm_id, "info", "approve", "Заявка одобрена")
			app.res.Body.Close()
		}
		rows.Close()
		datab.Close()
	}
}
