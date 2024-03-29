package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sbt/apay"
	"sbt/cusers"
	"sbt/logger"
	"sbt/payment"
	"sbt/random"
	"sbt/request"
	"strconv"
	"time"

	db "github.com/SashaShrek/db"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

var (
	ID_CHANNEL *int64
	ID_CHAT    *int64
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
			tgbotapi.NewKeyboardButton("ДОСТУП НА 1 МЕСЯЦ - 1399₽"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("КАК ОПЛАТИТЬ?"),
			tgbotapi.NewKeyboardButton("АВТОПЛАТЕЖИ"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Мой Telegram ID"),
		),
	)
)

type (
	tlgrm struct {
		count int
	}
	ResponseObj struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	ResponsePay struct {
		Event  string      `json:"event"`
		Object ResponseObj `json:"object"`
	}
	Data struct {
		tlgrm_id string
		MsgId    int
		PayId    string
	}
)

func readSecretData() {
	file, err := os.Open("data.xml")
	if err != nil {
		log := map[string]string{
			"User": "-1",
			"Func": "readSecretData",
		}
		logger.Take("error", log, err.Error())
		return
	}
	defer file.Close()

	type Database struct {
		Host     string `xml:"host"`
		Port     int    `xml:"port"`
		User     string `xml:"user"`
		Password string `xml:"password"`
		DBname   string `xml:"dbname"`
	}
	type Datas struct {
		XMLName    xml.Name `xml:"data"`
		Id_channel int64    `xml:"id_channel"`
		Id_chat    int64    `xml:"id_chat"`
		Price      string   `xml:"price"`
		ShopdId    string   `xml:"shopId"`
		BackLink   string   `xml:"backLink"`
		Token      string   `xml:"token"`
		PayToken   string   `xml:"payToken"`
		Datab      Database `xml:"database"`
	}
	var data Datas
	err = xml.NewDecoder(file).Decode(&data)
	if err != nil {
		log := map[string]string{
			"User": "-1",
			"Func": "readSecretData",
		}
		logger.Take("error", log, err.Error())
		return
	}
	ID_CHANNEL = &data.Id_channel
	ID_CHAT = &data.Id_chat
	PRICE = &data.Price
	SHOPID = &data.ShopdId
	BACK_LINK = &data.BackLink
	TOKEN = &data.Token
	PAY_TOKEN = &data.PayToken
	db.FillData(&data.Datab.Host, &data.Datab.Port, &data.Datab.User, &data.Datab.Password, &data.Datab.DBname)
	log := map[string]string{
		"User": "-1",
		"Func": "readSecretData",
	}
	logger.Take("info", log, "Data readed")
}

func main() {
	readSecretData()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET request")
		fmt.Fprintf(w, "Not found")
		log := map[string]string{
			"User": "-1",
			"Func": "main",
		}
		logger.Take("warn", log, "Get request")
	})
	botTime, err := createBotConnection(*TOKEN)
	if err != nil {
		logger.SetLog("-1", "error", "connectionBot", err.Error())
		log := map[string]string{
			"User": "-1",
			"Func": "main",
		}
		logger.Take("error", log, err.Error())
	}
	logger.SetLog("-1", "info", "connectionBot", "OK")
	log := map[string]string{
		"User": "-1",
		"Func": "main",
	}
	logger.Take("info", log, "Server to bot is connected")
	bot = botTime
	//go getCountUsers()
	//go newVersion()
	go approveInvite()
	go timer()
	go update()

	http.HandleFunc("/sbt_two_k_twenty_one", getResponse)
	http.ListenAndServe(":20021", nil)
}

func getCountUsers() {
	var query string
	var err error
	for range time.Tick(61 * time.Minute) {
		query = "insert into cusers (row_id, number, date) values " +
			"((select max(row_id) + 1 from cusers), (select count(*) from users where is_pay = true), current_timestamp)"
		err = db.InsertOrUpdate(query)
		if err != nil {
			logger.SetLog("-1", "error", "getCountUsers", err.Error())
			continue
		}
		logger.SetLog("-1", "info", "getCountUsers", "Успешно")
	}
}

func approveInvite() {
	cusers.Appr(bot, TOKEN, ID_CHANNEL)
}

func newVersion() {
	text := "Изменена стоимость подписки. Чтобы узнавать о технических работах (периоде их проведения), новшествах и исправлениях, можете стать участником специализированного для этих целей канала: https://t.me/+3MvQOOiNcZFiZmZi"
	rows, dbase, _ := db.Select("select tlgrm_id from users where is_pay = true")
	defer rows.Close()
	defer dbase.Close()
	type TlgrmId struct {
		tlgrm_id string
	}
	var tlgrmId TlgrmId
	var tgId int64
	var message tgbotapi.MessageConfig
	for rows.Next() {
		err := rows.Scan(&tlgrmId.tlgrm_id)
		if err != nil {
			continue
		}
		tgId, _ = strconv.ParseInt(tlgrmId.tlgrm_id, 10, 64)
		message = tgbotapi.NewMessage(tgId, text)
		message.ReplyMarkup = pay
		bot.Send(message)
	}
	logger.SetLog("-1", "info", "updated", "Сообщение всем отправлено")
}

func timer() {
	cusers.Keks(bot, TOKEN, ID_CHANNEL, ID_CHAT)
	for range time.Tick(24 * time.Hour) {
		cusers.Keks(bot, TOKEN, ID_CHANNEL, ID_CHAT)
	}
}

func getResponse(res http.ResponseWriter, req *http.Request) {
	logger.SetLog("-1", "info", "serverListen", "OK")
	if req.Method != "POST" {
		fmt.Fprintf(res, "Not found")
		log := map[string]string{
			"User": "-1",
			"Func": "getResponse",
		}
		logger.Take("warn", log, "Not POST")
	}

	var resp ResponsePay
	err := json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		log := map[string]string{
			"User": "-1",
			"Func": "getResponse",
		}
		logger.Take("error", log, err.Error())
	}

	var data Data
	rows, datab, err := db.Select(fmt.Sprintf("select tlgrm_id, message_pay_id, payment_id from users where id_last_transaction = '%s'", resp.Object.Id))
	if err != nil {
		log := map[string]string{
			"User": "-1",
			"Func": "getResponse",
		}
		logger.Take("error", log, err.Error())
		return
	}
	for rows.Next() {
		err := rows.Scan(&data.tlgrm_id, &data.MsgId, &data.PayId)
		if err != nil {
			continue
		}
	}
	tlgrm_id, err := strconv.ParseInt(data.tlgrm_id, 10, 64)
	//tlgrm_id, err := strconv.Atoi(data.tlgrm_id)
	if err != nil {
		log := map[string]string{
			"User": fmt.Sprint(tlgrm_id),
			"Func": "getResponse",
		}
		logger.Take("error", log, err.Error())
	}
	rows.Close()
	datab.Close()
	switch resp.Event {
	case "payment.succeeded":
		bot.Send(tgbotapi.NewMessage(tlgrm_id,
			payment.PaymentDone(tlgrm_id, resp.Object.Id, data.MsgId, TOKEN, ID_CHANNEL, ID_CHAT)))
		log := map[string]string{
			"User": fmt.Sprint(tlgrm_id),
			"Func": "getResponse",
		}
		logger.Take("info", log, "payment.succeeded")
	case "payment.canceled":
		log := map[string]string{
			"User": fmt.Sprint(tlgrm_id),
			"Func": "getResponse",
		}
		logger.Take("info", log, "payment.canceled")
		resp, _ := request.GetPaymentObj(*PRICE, *BACK_LINK, *SHOPID, *PAY_TOKEN, data.PayId)
		respUser, ownerUser, err := payment.PaymentCancel(resp.Cancel.Party, resp.Cancel.Reason, tlgrm_id)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(tlgrm_id, "Не удалось открыть файл конфигурации. Обратитесь к разработчику.\n\nВаш платеж был отменен!"))
			log := map[string]string{
				"User": fmt.Sprint(tlgrm_id),
				"Func": "getResponse",
			}
			logger.Take("error", log, err.Error())
		}
		log = map[string]string{
			"User": fmt.Sprint(tlgrm_id),
			"Func": "getResponse",
		}
		logger.Take("info", log, respUser)
		logger.SetLog(fmt.Sprint(tlgrm_id), "info", "payment", respUser)
		bot.Send(tgbotapi.NewMessage(tlgrm_id,
			fmt.Sprintf("Платёж был отменён %s! %s\nДля повторной оплаты нажмите кнопку оплатить внизу экрана.", ownerUser, respUser)))
	}
	bot.Send(tgbotapi.NewDeleteMessage(tlgrm_id, data.MsgId))
	err = db.InsertOrUpdate(fmt.Sprintf("update users set message_pay_id = 0, link_pay = null, id_last_transaction = null, payment_id = null where id_last_transaction = '%s'",
		resp.Object.Id))
	if err != nil {
		log := map[string]string{
			"User": fmt.Sprint(tlgrm_id),
			"Func": "getResponse",
		}
		logger.Take("error", log, err.Error())
	}
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
		log := map[string]string{
			"User": "-1",
			"Func": "update",
		}
		logger.Take("error", log, err.Error())
		logger.SetLog("-1", "error", "update", err.Error())
		return
	}
	logger.SetLog("-1", "info", "update", "OK")
	for update := range updates {
		if update.CallbackQuery != nil {
			var result bool
			var msgId int
			var message tgbotapi.CallbackConfig
			switch update.CallbackQuery.Data {
			case "/подключить":
				result, msgId = apay.GetAutoPay(update.CallbackQuery.From.ID)
				if result {
					message = tgbotapi.NewCallback(update.CallbackQuery.ID, "Автоплатёж уже подключен")
				} else {
					result = true
					err = db.InsertOrUpdate(fmt.Sprintf("update users set autopay = %t where tlgrm_id = '%s'",
						result, fmt.Sprint(update.CallbackQuery.From.ID)))
					if err != nil {
						log := map[string]string{
							"User": fmt.Sprint(update.CallbackQuery.From.ID),
							"Func": "update",
						}
						logger.Take("error", log, err.Error())
						return
					}
					message = tgbotapi.NewCallback(update.CallbackQuery.ID, "Автоплатёж успешно подключен")
				}
				message.ShowAlert = true
				bot.AnswerCallbackQuery(message)
				bot.Send(tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID),
					"Чтобы автоплатёж начал работать, необходимо провести оплату банковской картой!"))
				log := map[string]string{
					"User": fmt.Sprint(update.CallbackQuery.From.ID),
					"Func": "update",
				}
				logger.Take("info", log, "Autopay on")
			case "/отключить":
				result, msgId = apay.GetAutoPay(update.CallbackQuery.From.ID)
				if result {
					result = false
					err = db.InsertOrUpdate(fmt.Sprintf("update users set autopay = %t where tlgrm_id = '%s'",
						result, fmt.Sprint(update.CallbackQuery.From.ID)))
					if err != nil {
						log := map[string]string{
							"User": fmt.Sprint(update.CallbackQuery.From.ID),
							"Func": "update",
						}
						logger.Take("error", log, err.Error())
						return
					}
					message = tgbotapi.NewCallback(update.CallbackQuery.ID, "Автоплатёж успешно отключен")
				} else {
					message = tgbotapi.NewCallback(update.CallbackQuery.ID, "Автоплатёж уже отключен")
				}
				message.ShowAlert = true
				bot.AnswerCallbackQuery(message)
				log := map[string]string{
					"User": fmt.Sprint(update.CallbackQuery.From.ID),
					"Func": "update",
				}
				logger.Take("info", log, "Autopay off")
			}
			bot.Send(tgbotapi.NewDeleteMessage(int64(update.CallbackQuery.From.ID), msgId))
			err = db.InsertOrUpdate(fmt.Sprintf("update users set autopay_msg_id = 0 where tlgrm_id = '%s'",
				fmt.Sprint(update.CallbackQuery.From.ID)))
			if err != nil {
				log := map[string]string{
					"User": fmt.Sprint(update.CallbackQuery.From.ID),
					"Func": "update",
				}
				logger.Take("error", log, err.Error())
				return
			}
		}
		if update.Message == nil {
			continue
		}
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			if int64(*ID_CHAT) == update.Message.Chat.ID {
				continue
			}
			switch update.Message.Text {
			case "/start":
				msg := "Этот канал, создан для того, чтобы делиться с вами тем, что поможет сделать ваш гардероб стильным и красивым.\n" +
					"Лично помогать в выборе качественных вещей, разного сегмента и создании образов 🔥"
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
						log := map[string]string{
							"User": fmt.Sprint(update.Message.Chat.ID),
							"Func": "update",
						}
						logger.Take("error", log, err.Error())
					}
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "info", "start", "Пользователь добавлен в БД")
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "update",
					}
					logger.Take("info", log, "Added user")
				} else {
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "info", "start", "Пользователь уже существует в БД")
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "update",
					}
					logger.Take("info", log, "User already exists")
					//message = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вопросы по работе бота: supp.sbt@gmail.com\n\nВаш Telegram ID: %s, его необходимо указывать при каждом обращении на указанную почту.", fmt.Sprint(update.Message.Chat.ID)))
					//bot.Send(message)
				}
				rows.Close()
				datab.Close()
			case pay.Keyboard[0][0].Text:
				rand := random.GetRandom(18)
				res, bLink := request.GetPaymentObj(*PRICE, *BACK_LINK, *SHOPID, *PAY_TOKEN, rand) //08.06.2023 bLink - Запись полного ответа эквайринга
				/*08.06.2023 Запись полного ответа эквайринга*/
				var query string = fmt.Sprintf("insert into link_payment_obj (row_id, user_id, link_obj) values ('%s', (select row_id from users where tlgrm_id = '%s'), '%s')",
					random.GetRandom(10), fmt.Sprint(update.Message.Chat.ID), string(bLink))
				err = db.InsertOrUpdate(query)
				if err != nil {
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "insert",
					}
					logger.Take("error", log, err.Error())
				}
				var description string = "Онлайн шоппинг и мои личные рекомендации/секреты, будут доступны каждой из вас 💋"
				message := tgbotapi.NewMessage(update.Message.Chat.ID, description)
				message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf("Оплатить %s₽", *PRICE), res.ConfirmationsNew.ConfUrl),
					),
				)
				messageSended, err := bot.Send(message)
				if err != nil {
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "update",
					}
					logger.Take("error", log, err.Error())
					logger.SetLog(fmt.Sprint(update.Message.Chat.ID), "error", "getIdMsg", err.Error())
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				msgId := messageSended.MessageID
				query = fmt.Sprintf("update users set message_pay_id = %d, link_pay = '%s', id_last_transaction = '%s', payment_id = '%s' where tlgrm_id = '%s'",
					msgId, res.ConfirmationsNew.ConfUrl, res.Id, rand, fmt.Sprint(update.Message.Chat.ID))
				err = db.InsertOrUpdate(query)
				if err != nil {
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "update",
					}
					logger.Take("error", log, err.Error())
				}

			case pay.Keyboard[1][0].Text:
				howPayment := "Чтобы оплатить подписку, необходимо выполнить несколько простых шагов:\n" +
					"1. Нажать кнопку «ДОСТУП НА 1 МЕСЯЦ - 999₽»\n" +
					"2. Бот пришлет вам краткое описание услуги и кнопку «ОПЛАТИТЬ 999.00₽», необходимо нажать на неё\n" +
					"3. Подтвердите, при необходимости, открытие ссылки. Вы попадёте на страницу оплаты. На ней вы выбираете способ оплаты и вводите данные. Если необходима квитанция - поставьте галочку\n" +
					"4. После успешной оплаты бот пришлет вам ссылку-приглашение на канал\n" +
					"5. Чтобы вернуться обратно в чат к боту, вы можете нажать кнопку «Вернуться в магазин», потом нажать на зелёную кнопку в центре экрана («SEND MESSAGE») или просто закрыть данное окно.\n"
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, howPayment))
				log := map[string]string{
					"User": fmt.Sprint(update.Message.Chat.ID),
					"Func": "update",
				}
				logger.Take("info", log, "Instruction get")
			case pay.Keyboard[1][1].Text:
				text := "[ ТЕСТ - СЕЙЧАС АВТОПЛАТЕЖИ РАБОТАТЬ НЕ БУДУТ ]\nВы действительно хотите включить автоплатёж? Если хотите подключить - нажмите галочку (✓), если отключить - крестик (✗).\n" +
					"Автоплатёж можно подключить только к банковской карте!"
				message := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("✓", "/подключить"),
						tgbotapi.NewInlineKeyboardButtonData("✗", "/отключить"),
					),
				)
				msgId, _ := bot.Send(message)
				err = db.InsertOrUpdate(fmt.Sprintf("update users set autopay_msg_id = %d where tlgrm_id = '%s'",
					msgId.MessageID, fmt.Sprint(update.Message.Chat.ID)))
				if err != nil {
					log := map[string]string{
						"User": fmt.Sprint(update.Message.Chat.ID),
						"Func": "update",
					}
					logger.Take("error", log, err.Error())
				}
			case pay.Keyboard[2][0].Text:
				message := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ваш Telegram ID: %d", update.Message.Chat.ID))
				bot.Send(message)
			default:
				message := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет такой команды!")
				message.ReplyMarkup = pay
				bot.Send(message)
			}
		}
	}
}
