package cusers

import (
	"net/http"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type (
	IDates interface {
		Keks(*tgbotapi.BotAPI, *string, *int64, *int64)
	}
	IApprove interface {
		Appr(*tgbotapi.BotAPI, *string, *int64)
	}
)
type (
	dates struct {
		next_date_pay     time.Time
		notifier_date_pay time.Time
		is_pay            bool
		is_pay_first      bool
		tlgrm_id          string
	}
	appInv struct {
		tlgrm_id string
	}
	DataKick struct {
		dts     dates
		result  float64
		query   string
		dateNow time.Time
		message tgbotapi.MessageConfig
		err     error
		id      int64
		res     *http.Response
	}
	DataAprrove struct {
		result  bool
		desc    string
		id      int64
		aInv    appInv
		err     error
		data    []byte
		res     *http.Response
		message tgbotapi.MessageConfig
	}
)
