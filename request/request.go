package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sbt/logger"
)

type (
	CancelDetails struct {
		Party  string `json:"party"`
		Reason string `json:"reason"`
	}

	Confirmation struct {
		Types     string `json:"type"`
		ReturnURL string `json:"return_url"`
	}

	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	}

	Payload struct {
		Amounts       Amount       `json:"amount"`
		Capture       bool         `json:"capture"`
		Confirmations Confirmation `json:"confirmation"`
		Description   string       `json:"description"`
	}

	ConfirmationNew struct {
		Type    string `json:"type"`
		ConfUrl string `json:"confirmation_url"`
	}

	Response struct {
		Id               string          `json:"id"`
		Status           string          `json:"status"`
		ConfirmationsNew ConfirmationNew `json:"confirmation"`
		Cancel           CancelDetails   `json:"cancellation_details"`
	}
)

func GetPaymentObj(price string, url string, shopId string, token string, rand string) Response {
	body := bytes.NewReader(fillData(price, url))
	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", body)
	if err != nil {
		logger.SetLog("-1", "error", "GetPaymentObj", err.Error())
	}
	req.SetBasicAuth(shopId, token)
	req.Header.Set("Idempotence-Key", rand)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.SetLog("-1", "error", "GetPaymentObj, request", err.Error())
	}
	defer res.Body.Close()
	var resp Response
	json.NewDecoder(res.Body).Decode(&resp)
	return resp
}

func fillData(price string, url string) []byte {
	amountData := Amount{
		Value:    price,
		Currency: "RUB",
	}
	confirmData := Confirmation{
		Types:     "redirect",
		ReturnURL: url,
	}
	data := Payload{
		Amounts:       amountData,
		Confirmations: confirmData,
		Capture:       true,
		Description:   "Подписка на канал",
	}
	payBytes, err := json.Marshal(data)
	if err != nil {
		logger.SetLog("-1", "error", "fillData", err.Error())
	}
	return payBytes
}
