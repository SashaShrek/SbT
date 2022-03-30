package payment

import (
	"encoding/xml"
	"fmt"
	"os"
	"sbt/logger"
)

func PaymentCancel(ownerCancel string, reason string, id int64) (string, string, error) {
	type Reason struct {
		Name string `xml:"name"`
		Text string `xml:"text"`
	}

	type Owner struct {
		Name string `xml:"name"`
		Text string `xml:"text"`
	}

	type TypePC struct {
		XML  xml.Name `xml:"type_pc"`
		Rsn  []Reason `xml:"reasons>reason"`
		Ownr []Owner  `xml:"owners>owner"`
	}

	sid := fmt.Sprint(id)
	file, err := os.Open("payment_cancel.xml")
	if err != nil {
		logger.SetLog(sid, "error", "paymentCancel", err.Error())
		return "", "", err
	}
	defer file.Close()
	var typePC TypePC
	_ = xml.NewDecoder(file).Decode(&typePC)
	var resReason, resOwner string = "0", "0"
	for _, item := range typePC.Rsn {
		if reason == item.Name {
			resReason = item.Text
			break
		}
	}
	for _, item := range typePC.Ownr {
		if ownerCancel == item.Name {
			resOwner = item.Text
			break
		}
	}
	if resReason == "0" {
		resReason = "Неизвестный статус"
	}
	if resOwner == "0" {
		resOwner = "неизвестно кем"
	}
	return resReason, resOwner, nil
}
