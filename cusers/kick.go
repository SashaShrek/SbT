package cusers

import (
	"fmt"
	"net/http"
	"sbt/logger"
	"strconv"
	"time"

	"github.com/SashaShrek/db"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

func (d DataKick) Keks(bot *tgbotapi.BotAPI, token *string, idChannel *int64, idChat *int64) {
	d.query = "select next_date_pay, notifier_date_pay, is_pay, is_pay_first, tlgrm_id from users where is_pay = true"
	rows, datab, _ := db.Select(d.query)
	d.dateNow = time.Now()
	for rows.Next() {
		d.err = rows.Scan(&d.dts.next_date_pay, &d.dts.notifier_date_pay, &d.dts.is_pay,
			&d.dts.is_pay_first, &d.dts.tlgrm_id)
		if d.err != nil {
			continue
		}
		if d.dts.is_pay_first {
			if !d.dts.is_pay {
				continue
			}
			if d.dateNow.Sub(d.dts.next_date_pay).Hours() > 24 {
				d.dts.is_pay = false
				d.err = db.InsertOrUpdate(fmt.Sprintf("update users set is_pay = %t where tlgrm_id = '%s'", d.dts.is_pay, d.dts.tlgrm_id))
				if d.err != nil {
					fmt.Println(d.err)
				}
			}
			if !d.dts.is_pay {
				d.res, d.err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/banChatMember?chat_id=%d&user_id=%s", *token, *idChannel, d.dts.tlgrm_id))
				if d.err != nil {
					logger.SetLog(d.dts.tlgrm_id, "error", "banUser", d.err.Error())
				}
				d.res.Body.Close()
				d.res, d.err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/banChatMember?chat_id=%d&user_id=%s", *token, *idChat, d.dts.tlgrm_id))
				if d.err != nil {
					logger.SetLog(d.dts.tlgrm_id, "error", "banUserFromChat", d.err.Error())
				}
				d.res.Body.Close()
				/*res, err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember?chat_id=%d&user_id=%s", *TOKEN, *ID_CHAT, dates.tlgrm_id))
				if err != nil {
					logger.SetLog(dates.tlgrm_id, "error", "unbanUserFromChat", err.Error())
				}
				res.Body.Close()*/

				logger.SetLog(d.dts.tlgrm_id, "info", "banUser", "Кикнут")
				d.id, _ = strconv.ParseInt(d.dts.tlgrm_id, 10, 64)
				d.message = tgbotapi.NewMessage(d.id, "Доступ к каналу STYLE by Tsymlyanskaya отозван")
				bot.Send(d.message)
				continue
			}
			d.result = d.dateNow.Sub(d.dts.notifier_date_pay).Hours()
			if (d.result >= 0 && d.result < 24) || (d.result >= 0 && d.result >= 48) {
				d.id, _ = strconv.ParseInt(d.dts.tlgrm_id, 10, 64)
				d.message = tgbotapi.NewMessage(d.id, "Необходимо продлить подписку. В противном случае доступ к каналу STYLE by Tsymlyanskaya будет отозван!\nСделать это вы можете, нажав кнопку внизу экрана")
				bot.Send(d.message)
				logger.SetLog(d.dts.tlgrm_id, "info", "warnPay", "Напоминание о платеже отправлено")
			}
		}
	}
	rows.Close()
	datab.Close()
}
