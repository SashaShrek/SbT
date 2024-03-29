package payment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sbt/logger"
	"time"

	"github.com/SashaShrek/db"
)

type (
	IsPay struct {
		is_pay       bool
		is_pay_first bool
	}

	DatesPay struct {
		date_p     time.Time
		n_date_p   time.Time
		not_date_p time.Time
	}
)

func PaymentDone(tlgrm_id int64, transaction string, msgId int, token *string, idChannel *int64, idChat *int64) string {
	stg_id := fmt.Sprint(tlgrm_id)
	rows, datab, err := db.Select(fmt.Sprintf("select is_pay, is_pay_first from users where tlgrm_id = '%s'",
		stg_id))
	if err != nil {
		log := map[string]string{
			"User": stg_id,
			"Func": "PaymentDone",
		}
		logger.Take("error", log, err.Error())
	}
	var isPay IsPay
	for rows.Next() {
		err := rows.Scan(&isPay.is_pay, &isPay.is_pay_first)
		if err != nil {
			continue
		}
	}
	rows.Close()
	datab.Close()
	var msg string
	var invite interface{}
	if !isPay.is_pay && !isPay.is_pay_first {
		res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/createChatInviteLink?chat_id=%d&creates_join_request=true", *token, *idChannel))
		if err != nil {
			log := map[string]string{
				"User": stg_id,
				"Func": "PaymentDone",
			}
			logger.Take("error", log, err.Error())
			msg = fmt.Sprintf("Произошла ошибка: %s", err.Error())
			logger.SetLog(stg_id, "error", "createLink", err.Error())
			rows.Close()
			datab.Close()
			return msg
		}
		var dataRes map[string]interface{}
		data, _ := ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(data, &dataRes)
		invite = dataRes["result"].(map[string]interface{})["invite_link"].(string)
		res.Body.Close()
		msg = fmt.Sprintf("Подписка оплачена!\nПерейдите по ссылке и подайте заявку на вступление, чтобы получить доступ к каналу. Доступ будет предоставлен в течении 15 минут, если заявка была подана вами, а не 3-им лицом: %s",
			invite)
	} else if !isPay.is_pay {
		rows, datab, _ = db.Select(fmt.Sprintf("select link from users where tlgrm_id = '%s'",
			stg_id))
		type LinkUser struct {
			link string
		}
		var linkU LinkUser
		for rows.Next() {
			err := rows.Scan(&linkU.link)
			if err != nil {
				continue
			}
		}
		rows.Close()
		datab.Close()
		msg = fmt.Sprintf("Подписка оплачена!\nПерейдите по ссылке и подайте заявку на вступление, чтобы получить доступ к каналу. Доступ будет предоставлен в течении 15 минут, если заявка была подана вами, а не 3-им лицом: %s", linkU.link)
		log := map[string]string{
			"User": stg_id,
			"Func": "PaymentDone",
		}
		logger.Take("info", log, "Payment OK")
	} else if isPay.is_pay {
		msg = "Подписка продлена!"
	}

	var datesP DatesPay
	rows, datab, err = db.Select(fmt.Sprintf("select date_pay, next_date_pay, notifier_date_pay from users where tlgrm_id = '%s'", stg_id))
	if err != nil {
		log := map[string]string{
			"User": stg_id,
			"Func": "PaymentDone",
		}
		logger.Take("error", log, err.Error())
	}
	for rows.Next() {
		err := rows.Scan(&datesP.date_p, &datesP.n_date_p, &datesP.not_date_p)
		if err != nil {
			//logger.SetLog(stg_id, "error", "payment", err.Error())
			continue
		}
	}
	rows.Close()
	datab.Close()
	date_pay := time.Now()
	var next_date_pay time.Time
	var notifier_date_pay time.Time
	delta := datesP.n_date_p.Sub(date_pay)
	if delta > 0 {
		next_date_pay = datesP.n_date_p.AddDate(0, 1, 0)
		notifier_date_pay = datesP.n_date_p.AddDate(0, 1, -2)
	} else {
		next_date_pay = date_pay.AddDate(0, 1, 0)
		notifier_date_pay = date_pay.AddDate(0, 1, -2)
	}
	var query string
	if !isPay.is_pay && !isPay.is_pay_first {
		query = fmt.Sprintf("update users set link = '%s', is_pay = true, is_pay_first = true, invite = true,"+
			"date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"next_date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"notifier_date_pay = to_date('%s', 'YYYY-MM-DD') where tlgrm_id = '%s'",
			invite, date_pay, next_date_pay, notifier_date_pay, stg_id)
	} else if !isPay.is_pay {
		query = fmt.Sprintf("update users set is_pay = true, is_pay_first = true, invite = true,"+
			"date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"next_date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"notifier_date_pay = to_date('%s', 'YYYY-MM-DD') where tlgrm_id = '%s'",
			date_pay, next_date_pay, notifier_date_pay, stg_id)
		res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember?chat_id=%d&user_id=%d", *token, *idChannel, tlgrm_id))
		if err != nil {
			log := map[string]string{
				"User": stg_id,
				"Func": "PaymentDone",
			}
			logger.Take("error", log, err.Error())
			logger.SetLog(stg_id, "error", "unbanUser", err.Error())
		}
		res.Body.Close()
		res, err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember?chat_id=%d&user_id=%d", *token, *idChat, tlgrm_id))
		if err != nil {
			log := map[string]string{
				"User": stg_id,
				"Func": "PaymentDone",
			}
			logger.Take("error", log, err.Error())
			logger.SetLog(stg_id, "error", "unbanUser", err.Error())
		}
		res.Body.Close()
	} else if isPay.is_pay {
		query = fmt.Sprintf("update users set is_pay = true, is_pay_first = true,"+
			"date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"next_date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"notifier_date_pay = to_date('%s', 'YYYY-MM-DD') where tlgrm_id = '%s'",
			date_pay, next_date_pay, notifier_date_pay, stg_id)
	}
	err = db.InsertOrUpdate(query)
	if err != nil {
		log := map[string]string{
			"User": stg_id,
			"Func": "PaymentDone",
		}
		logger.Take("error", log, err.Error())
		logger.SetLog(stg_id, "error", "payment", err.Error())
	}

	query = fmt.Sprintf("insert into transaction (row_id, provider_token_payment) values ((select row_id from users where tlgrm_id = '%s'), '%s')",
		stg_id, transaction)
	err = db.InsertOrUpdate(query)
	if err != nil {
		log := map[string]string{
			"User": stg_id,
			"Func": "PaymentDone",
		}
		logger.Take("error", log, err.Error())
	}
	logger.SetLog(stg_id, "info", "payment", "Оплата прошла успешно")
	return msg
}
