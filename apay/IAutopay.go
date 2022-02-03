package apay

type (
	IAPay interface {
		GetAutoPay(id int) (bool, int)
	}

	dataPay struct {
		Autopay bool
		MsgId   int
	}

	Data struct {
		dtPay       dataPay
		requestData string
	}
)
