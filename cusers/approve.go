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

func (app DataAprrove) Appr(bot *tgbotapi.BotAPI, token *string, idChannel *int64) {
	for range time.Tick(15 * time.Minute) {
		rows, datab, _ := db.Select("select tlgrm_id from users where invite = true")
		for rows.Next() {
			app.err = rows.Scan(&app.aInv.tlgrm_id)
			if app.err != nil {
				continue
			}
			app.id, _ = strconv.ParseInt(app.aInv.tlgrm_id, 10, 64)
			app.res, app.err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/approveChatJoinRequest?chat_id=%d&user_id=%s", *token, *idChannel, app.aInv.tlgrm_id))
			if app.err != nil {
				logger.SetLog(app.aInv.tlgrm_id, "error", "approve", app.err.Error())
				continue
			}
			var dataRes map[string]interface{}
			app.data, _ = ioutil.ReadAll(app.res.Body)
			_ = json.Unmarshal(app.data, &dataRes)

			app.result = dataRes["ok"].(bool)
			if dataRes["description"] != nil {
				app.desc = dataRes["description"].(string)
			}
			if !app.result {
				logger.SetLog(app.aInv.tlgrm_id, "warn", "approve", app.desc)
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
