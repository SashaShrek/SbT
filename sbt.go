package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sbt/logger"
	"sbt/random"
	"sbt/request"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	db "github.com/SashaShrek/db"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

var (
	ID_CHANNEL *int
	ID_CHAT    *int
	PRICE      *string
	SHOPID     *string
	BACK_LINK  *string
	TOKEN      *string
	PAY_TOKEN  *string
)

var (
	bot *tgbotapi.BotAPI
	pay = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ДОСТУП НА 1 МЕСЯЦ - 999₽"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("КАК ОПЛАТИТЬ?"),
		),
	)
)

type (
	Dates struct {
		next_date_pay     time.Time
		notifier_date_pay time.Time
		is_pay            bool
		is_pay_first      bool
		tlgrm_id          string
	}

	IsPay struct {
		is_pay       bool
		is_pay_first bool
	}

	DatesPay struct {
		date_p     time.Time
		n_date_p   time.Time
		not_date_p time.Time
	}

	tlgrm struct {
		count int
	}
)

func readSecretData() {
	file, err := os.Open("data.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	type Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBname   string `yaml:"dbname"`
	}
	type Data struct {
		Id_channel int      `yaml:"id_channel"`
		Id_chat    int      `yaml:"id_chat"`
		Price      string   `yaml:"price"`
		ShopdId    string   `yaml:"shopId"`
		BackLink   string   `yaml:"backLink"`
		Token      string   `yaml:"token"`
		PayToken   string   `yaml:"payToken"`
		Datab      Database `yaml:"database"`
	}
	var data Data
	_ = yaml.NewDecoder(file).Decode(&data)
	ID_CHANNEL = &data.Id_channel
	ID_CHAT = &data.Id_chat
	PRICE = &data.Price
	SHOPID = &data.ShopdId
	BACK_LINK = &data.BackLink
	TOKEN = &data.Token
	PAY_TOKEN = &data.PayToken
	db.FillData(&data.Datab.Host, &data.Datab.Port, &data.Datab.User, &data.Datab.Password, &data.Datab.DBname)
}

func main() {
	readSecretData()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET запрос")
		fmt.Fprintf(w, "Not found")
	})
	botTime, err := createBotConnection(*TOKEN)
	if err != nil {
		logger.SetLog("-1", "error", "connectionBot", err.Error())
	}
	logger.SetLog("-1", "info", "connectionBot", "OK")
	bot = botTime
	go newVersion()
	go approveInvite()
	go timer()
	go update()

	http.HandleFunc("/sbt_two_k_twenty_one", getResponse)
	http.ListenAndServe(":20021", nil)
}

func approveInvite() {
	for range time.Tick(15 * time.Minute) {
		rows, datab, _ := db.Select("select tlgrm_id from users where invite = true")
		type AppInv struct {
			tlgrm_id string
		}
		var appInv AppInv
		for rows.Next() {
			err := rows.Scan(&appInv.tlgrm_id)
			if err != nil {
				continue
			}
			res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/approveChatJoinRequest?chat_id=%d&user_id=%s", *TOKEN, *ID_CHANNEL, appInv.tlgrm_id))
			if err != nil {
				logger.SetLog(appInv.tlgrm_id, "error", "approve", err.Error())
			}
			var dataRes map[string]interface{}
			data, _ := ioutil.ReadAll(res.Body)
			_ = json.Unmarshal(data, &dataRes)
			result := dataRes["ok"]
			if result != true {
				res.Body.Close()
				continue
			}
			_ = db.InsertOrUpdate(fmt.Sprintf("update users set invite = false where tlgrm_id = '%s'", appInv.tlgrm_id))
			id, _ := strconv.ParseInt(appInv.tlgrm_id, 10, 64)
			message := tgbotapi.NewMessage(id, "Доступ к каналу предоставлен: заявка одобрена")
			bot.Send(message)
			logger.SetLog(appInv.tlgrm_id, "info", "approve", "Заявка одобрена")
			res.Body.Close()
		}
		rows.Close()
		datab.Close()
	}
}

func newVersion() {
	text := "Что нового в версии 1.2.0:\n" +
		"- Полностью изменена процедура оплаты: теперь для вступления на канал требуется подать заявку на вступление\n" +
		"- Исправлены некоторые ошибки, связанные с интеграцией c базой данных."
	rows, dbase, _ := db.Select("select tlgrm_id from users")
	defer rows.Close()
	defer dbase.Close()
	type TlgrmId struct {
		tlgrm_id string
	}
	var tlgrmId TlgrmId
	var tgId int64
	for rows.Next() {
		err := rows.Scan(&tlgrmId.tlgrm_id)
		if err != nil {
			continue
		}
		tgId, _ = strconv.ParseInt(tlgrmId.tlgrm_id, 10, 64)
		message := tgbotapi.NewMessage(tgId, text)
		message.ReplyMarkup = pay
		bot.Send(message)
	}
	logger.SetLog("-1", "info", "updated", "Сообщение всем отправлено")
}

func kicker() {
	var result float64
	var dates Dates
	var query string = "select next_date_pay, notifier_date_pay, is_pay, is_pay_first, tlgrm_id from users"
	rows, datab, _ := db.Select(query)
	dateNow := time.Now()
	for rows.Next() {
		err := rows.Scan(&dates.next_date_pay, &dates.notifier_date_pay, &dates.is_pay,
			&dates.is_pay_first, &dates.tlgrm_id)
		if err != nil {
			continue
		}
		if dates.is_pay_first {
			if !dates.is_pay {
				continue
			}
			if dateNow.Sub(dates.next_date_pay).Hours() > 24 {
				dates.is_pay = false
				err = db.InsertOrUpdate(fmt.Sprintf("update users set is_pay = %t where tlgrm_id = '%s'", dates.is_pay, dates.tlgrm_id))
				if err != nil {
					fmt.Println(err)
				}
			}
			if !dates.is_pay {
				res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/banChatMember?chat_id=%d&user_id=%s", *TOKEN, *ID_CHANNEL, dates.tlgrm_id))
				if err != nil {
					logger.SetLog(dates.tlgrm_id, "error", "banUser", err.Error())
				}
				res.Body.Close()
				res, err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/banChatMember?chat_id=%d&user_id=%s", *TOKEN, *ID_CHAT, dates.tlgrm_id))
				if err != nil {
					logger.SetLog(dates.tlgrm_id, "error", "banUserFromChat", err.Error())
				}
				res.Body.Close()
				res, err = http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember?chat_id=%d&user_id=%s", *TOKEN, *ID_CHAT, dates.tlgrm_id))
				if err != nil {
					logger.SetLog(dates.tlgrm_id, "error", "unbanUserFromChat", err.Error())
				}
				res.Body.Close()
				logger.SetLog(dates.tlgrm_id, "info", "banUser", "Кикнут")
				id, _ := strconv.ParseInt(dates.tlgrm_id, 10, 64)
				message := tgbotapi.NewMessage(id, "Доступ к каналу STYLE by Tsymlyanskaya отозван")
				bot.Send(message)
				continue
			}
			result = dateNow.Sub(dates.notifier_date_pay).Hours()
			if (result >= 0 && result < 24) || (result >= 0 && result >= 48) {
				id, _ := strconv.ParseInt(dates.tlgrm_id, 10, 64)
				message := tgbotapi.NewMessage(id, "Необходимо продлить подписку. В противном случае доступ к каналу STYLE by Tsymlyanskaya будет отозван!\nСделать это вы можете, нажав кнопку внизу экрана")
				bot.Send(message)
				logger.SetLog(dates.tlgrm_id, "info", "warnPay", "Напоминание о платеже отправлено")
			}
		}
	}
	rows.Close()
	datab.Close()
}

func timer() {
	kicker()
	for range time.Tick(24 * time.Hour) {
		kicker()
	}
}

func getResponse(res http.ResponseWriter, req *http.Request) {
	logger.SetLog("-1", "info", "serverListen", "OK")
	if req.Method != "POST" {
		fmt.Fprintf(res, "Not found")
	}

	type ResponseObj struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	type ResponsePay struct {
		Event  string      `json:"event"`
		Object ResponseObj `json:"object"`
	}
	var resp ResponsePay
	err := json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		fmt.Println(err)
	}

	type Data struct {
		tlgrm_id string
		MsgId    int
		PayId    string
	}
	var data Data
	rows, datab, _ := db.Select(fmt.Sprintf("select tlgrm_id, message_pay_id, payment_id from users where id_last_transaction = '%s'", resp.Object.Id))
	for rows.Next() {
		err := rows.Scan(&data.tlgrm_id, &data.MsgId, &data.PayId)
		if err != nil {
			continue
		}
	}
	tlgrm_id, _ := strconv.ParseInt(data.tlgrm_id, 10, 64)
	rows.Close()
	datab.Close()
	switch resp.Event {
	case "payment.succeeded":
		paymentDone(tlgrm_id, resp.Object.Id, data.MsgId)
	case "payment.canceled":
		resp := request.GetPaymentObj(*PRICE, *BACK_LINK, *SHOPID, *PAY_TOKEN, data.PayId)
		respUser, ownerUser := paymentCancel(resp.Cancel.Party, resp.Cancel.Reason, tlgrm_id)
		logger.SetLog(fmt.Sprint(tlgrm_id), "warn", "payment", respUser)
		response := fmt.Sprintf("Платёж был отменён %s! %s\nДля повторной оплаты нажмите кнопку оплатить внизу экрана.", ownerUser, respUser)
		bot.Send(tgbotapi.NewMessage(tlgrm_id, response))
	}
	bot.Send(tgbotapi.NewDeleteMessage(tlgrm_id, data.MsgId))
	_ = db.InsertOrUpdate(fmt.Sprintf("update users set message_pay_id = 0, link_pay = null, id_last_transaction = null, payment_id = null where id_last_transaction = '%s'",
		resp.Object.Id))
}
func paymentCancel(ownerCancel string, reason string, id int64) (string, string) {
	var responseUser string
	switch reason {
	case "3d_secure_failed":
		responseUser = "Не пройдена аутентификация по 3-D Secure."
	case "call_issuer":
		responseUser = "Оплата данным платежным средством отклонена по неизвестным причинам. Пользователю следует обратиться в организацию, выпустившую платежное средство."
	case "card_expired":
		responseUser = "Истек срок действия банковской карты."
	case "country_forbidden":
		responseUser = "Нельзя заплатить банковской картой, выпущенной в этой стране."
	case "expired_on_confirmation":
		responseUser = "Истек срок оплаты: пользователь не подтвердил платеж за время, отведенное на оплату выбранным способом. Если вы успешно оплатили, то проигнорируйте данное сообщение."
	case "fraud_suspected":
		responseUser = "Платеж заблокирован из-за подозрения в мошенничестве."
	case "general_decline":
		responseUser = "Причина не детализирована. Пользователю следует обратиться к инициатору отмены платежа за уточнением подробностей."
	case "identification_required":
		responseUser = "Превышены ограничения на платежи для кошелька ЮMoney."
	case "insufficient_funds":
		responseUser = "Не хватает денег для оплаты."
	case "internal_timeout":
		responseUser = "Технические неполадки на стороне ЮKassa: не удалось обработать запрос в течение 30 секунд."
	case "invalid_card_number":
		responseUser = "Неправильно указан номер карты."
	case "invalid_csc":
		responseUser = "Неправильно указан код CVV2 (CVC2, CID)."
	case "issuer_unavailable":
		responseUser = "Организация, выпустившая платежное средство, недоступна."
	case "payment_method_limit_exceeded":
		responseUser = "Исчерпан лимит платежей для данного платежного средства или вашего магазина."
	case "payment_method_restricted":
		responseUser = "Запрещены операции данным платежным средством (например, карта заблокирована из-за утери, кошелек — из-за взлома мошенниками). Пользователю следует обратиться в организацию, выпустившую платежное средство."
	case "permission_revoked":
		responseUser = "Нельзя провести безакцептное списание: пользователь отозвал разрешение на автоплатежи."
	case "unsupported_mobile_operator":
		responseUser = "Нельзя заплатить с номера телефона этого мобильного оператора. Доступные операторы: Мегафон, Билайн, МТС, Теле2"
	default:
		responseUser = "Неизвестный статус"
	}

	var owner string
	switch ownerCancel {
	case "merchant":
		owner = "продавцом товаров и услуг"
	case "yoo_money":
		owner = "ЮKassa"
	case "payment_network":
		owner = "«внешними» участниками платежного процесса — все остальные участники платежного процесса, кроме ЮKassa и продавца (например, эмитент, сторонний платежный сервис)"
	default:
		owner = "неизвестно кем"
	}
	return responseUser, owner
}
func paymentDone(tlgrm_id int64, transaction string, msgId int) {
	rows, datab, _ := db.Select(fmt.Sprintf("select is_pay, is_pay_first from users where tlgrm_id = '%s'",
		fmt.Sprint(tlgrm_id)))
	var isPay IsPay
	for rows.Next() {
		err := rows.Scan(&isPay.is_pay, &isPay.is_pay_first)
		if err != nil {
			continue
		}
	}
	var msg string
	var invite interface{}
	if !isPay.is_pay && !isPay.is_pay_first {
		res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/createChatInviteLink?chat_id=%d&creates_join_request=true", *TOKEN, *ID_CHANNEL))
		if err != nil {
			msg := fmt.Sprintf("Произошла ошибка: %s. Обратитесь в тех. поддержку по адресу supp.sbt@gmail.com", err.Error())
			message := tgbotapi.NewMessage(tlgrm_id, msg)
			bot.Send(message)
			logger.SetLog(fmt.Sprint(tlgrm_id), "error", "createLink", err.Error())
		}
		var dataRes map[string]interface{}
		data, _ := ioutil.ReadAll(res.Body)
		_ = json.Unmarshal(data, &dataRes)
		invite = dataRes["result"].(map[string]interface{})["invite_link"]
		res.Body.Close()
		msg = fmt.Sprintf("Подписка оплачена!\nПерейдите по ссылке и подайте заявку на вступление, чтобы получить доступ к каналу. Доступ будет предоставлен в течении 15 минут, если заявка была подана вами, а не 3-им лицом: %s\n\nВопросы по работе бота: supp.sbt@gmail.com\n\nВаш Telegram ID: %s, его необходимо указывать при каждом обращении на указанную почту.",
			invite, fmt.Sprint(tlgrm_id))
	} else if !isPay.is_pay {
		rows, datab, _ = db.Select(fmt.Sprintf("select link from users where tlgrm_id = '%s'",
			fmt.Sprint(tlgrm_id)))
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
		msg = fmt.Sprintf("Подписка оплачена!\nПерейдите по ссылке и подайте заявку на вступление, чтобы получить доступ к каналу. Доступ будет предоставлен в течении 15 минут, если заявка была подана вами, а не 3-им лицом: %s\n\nВопросы по работе бота: supp.sbt@gmail.com", linkU.link)
	} else if isPay.is_pay {
		msg = "Подписка продлена!\n\nВопросы по работе бота: supp.sbt@gmail.com"
	}
	rows.Close()
	datab.Close()
	message := tgbotapi.NewMessage(tlgrm_id, msg)
	bot.Send(message)

	var datesP DatesPay
	rows, datab, _ = db.Select(fmt.Sprintf("select date_pay, next_date_pay, notifier_date_pay from users where tlgrm_id = '%s'", fmt.Sprint(tlgrm_id)))
	for rows.Next() {
		err := rows.Scan(&datesP.date_p, &datesP.n_date_p, &datesP.not_date_p)
		if err != nil {
			fmt.Println(err.Error())
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
			invite, date_pay, next_date_pay, notifier_date_pay, fmt.Sprint(tlgrm_id))
	} else if !isPay.is_pay {
		query = fmt.Sprintf("update users set is_pay = true, is_pay_first = true, invite = true"+
			"date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"next_date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"notifier_date_pay = to_date('%s', 'YYYY-MM-DD') where tlgrm_id = '%s'",
			date_pay, next_date_pay, notifier_date_pay, fmt.Sprint(tlgrm_id))
		res, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/unbanChatMember?chat_id=%d&user_id=%s", *TOKEN, *ID_CHANNEL, fmt.Sprint(tlgrm_id)))
		if err != nil {
			logger.SetLog(fmt.Sprint(tlgrm_id), "error", "unbanUser", err.Error())
		}
		res.Body.Close()
	} else if isPay.is_pay {
		query = fmt.Sprintf("update users set is_pay = true, is_pay_first = true,"+
			"date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"next_date_pay = to_date('%s', 'YYYY-MM-DD'), "+
			"notifier_date_pay = to_date('%s', 'YYYY-MM-DD') where tlgrm_id = '%s'",
			date_pay, next_date_pay, notifier_date_pay, fmt.Sprint(tlgrm_id))
	}
	err := db.InsertOrUpdate(query)
	if err != nil {
		fmt.Println(err)
	}

	query = fmt.Sprintf("insert into transaction (row_id, provider_token_payment) values ((select row_id from users where tlgrm_id = '%s'), '%s')",
		fmt.Sprint(tlgrm_id), transaction)
	_ = db.InsertOrUpdate(query)
	logger.SetLog(fmt.Sprint(tlgrm_id), "info", "payment", "Оплата прошла успешно")
}

func createBotConnection(token string) (*tgbotapi.BotAPI, error) {
	botConnect, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return botConnect, nil
}

func update() {
	bot.Debug = false
	var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := bot.GetUpdatesChan(ucfg)
	if err != nil {
		logger.SetLog("-1", "error", "update", err.Error())
		return
	}
	logger.SetLog("-1", "info", "update", "OK")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			if int64(*ID_CHAT) == update.Message.Chat.ID {
				continue
			}
			switch update.Message.Text {
			case "/start":
				msg := "♥️Этот канал, создан для того, чтобы делиться с вами всем, что встречается мне на разных" +
					" площадках с прямыми ссылками на вещи, стоимостью и голосовыми сопровождениями. " +
					"У вас будет возможность, выбирать для себя только лучшие вещи разного сегмента, покупать бренды" +
					" и винтаж по цене Масс-маркета, иметь меня всегда рядом, как личного стилиста и подругу, которая" +
					" плохого не посоветует 😈"
				message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
				bot.Send(message)
				msg = fmt.Sprintf("Привет, %s!\nНажав на кнопку внизу экрана,"+
					" ты можешь оформить подписку на канал STYLE by Tsymlyanskaya.", update.Message.From.FirstName)
				message = tgbotapi.NewMessage(update.Message.Chat.ID, msg)
				message.ReplyMarkup = pay
				bot.Send(message)
				var query string = fmt.Sprintf("select count(*) from users where tlgrm_id = '%s'", fmt.Sprint(update.Message.Chat.ID))
				rows, datab, err := db.Select(query)
				if err != nil {
					logger.SetLog("null", "error", "start", err.Error())
					rows.Close()
					datab.Close()
					continue
				}
				var id tlgrm
				for rows.Next() {
					err = rows.Scan(&id.count)
					if err != nil {
						continue
					}
				}
				if id.count == 0 {
					query = fmt.Sprintf("insert into users (row_id, tlgrm_id"+
						") values ('%s', '%s')", random.GetRandom(10), fmt.Sprint(update.Message.Chat.ID))
					err = db.InsertOrUpdate(query)
					if err != nil {
						fmt.Println(err)
					}
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "info", "start", "Пользователь добавлен в БД")
				} else {
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "info", "start", "Пользователь уже существует в БД")
					message = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вопросы по работе бота: supp.sbt@gmail.com\n\nВаш Telegram ID: %s, его необходимо указывать при каждом обращении на указанную почту.", fmt.Sprint(update.Message.Chat.ID)))
					bot.Send(message)
				}
				rows.Close()
				datab.Close()
			case pay.Keyboard[0][0].Text:
				rand := random.GetRandom(18)
				res := request.GetPaymentObj(*PRICE, *BACK_LINK, *SHOPID, *PAY_TOKEN, rand)
				var description string = "Онлайн шоппинг и мои личные рекомендации/секреты, будут доступны каждой из вас 💋"
				message := tgbotapi.NewMessage(update.Message.Chat.ID, description)
				message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf("Оплатить %s₽", *PRICE), res.ConfirmationsNew.ConfUrl),
					),
				)
				messageSended, err := bot.Send(message)
				if err != nil {
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "error", "getIdMsg", err.Error())
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				msgId := messageSended.MessageID
				var query string = fmt.Sprintf("update users set message_pay_id = %d, link_pay = '%s', id_last_transaction = '%s', payment_id = '%s' where tlgrm_id = '%s'",
					msgId, res.ConfirmationsNew.ConfUrl, res.Id, rand, fmt.Sprint(update.Message.Chat.ID))
				_ = db.InsertOrUpdate(query)
			case pay.Keyboard[1][0].Text:
				howPayment := "Чтобы оплатить подписку, необходимо выполнить несколько простых шагов:\n" +
					"1. Нажать кнопку «ДОСТУП НА 1 МЕСЯЦ - 999₽»\n" +
					"2. Бот пришлет вам краткое описание услуги и кнопку «ОПЛАТИТЬ 999.00₽», необходимо нажать на неё\n" +
					"3. Подтвердите, при необходимости, открытие ссылки. Вы попадёте на страницу оплаты. На ней вы выбираете способ оплаты и вводите данные. Если необходима квитанция - поставьте галочку\n" +
					"4. После успешной оплаты бот пришлет вам ссылку-приглашение на канал\n" +
					"5. Чтобы вернуться обратно в чат к боту, вы можете нажать кнопку «Вернуться в магазин», потом нажать на зелёную кнопку в центре экрана («SEND MESSAGE») или просто закрыть данное окно.\n"
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, howPayment))
			default:
				message := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет такой команды!")
				message.ReplyMarkup = pay
				bot.Send(message)
			}
		}
	}
}
