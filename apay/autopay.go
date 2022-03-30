package apay

import (
	"fmt"

	"github.com/SashaShrek/db"
)

func GetAutoPay(id int) (bool, int) {
	type dataPay struct {
		Autopay bool
		MsgId   int
	}

	type Data struct {
		dtPay       dataPay
		requestData string
	}

	var d Data

	d.requestData = fmt.Sprintf("select autopay, autopay_msg_id from users where tlgrm_id = '%s'",
		fmt.Sprint(id))
	rows, datab, _ := db.Select(d.requestData)
	for rows.Next() {
		err := rows.Scan(&d.dtPay.Autopay, &d.dtPay.MsgId)
		if err != nil {
			continue
		}
	}
	rows.Close()
	datab.Close()
	return d.dtPay.Autopay, d.dtPay.MsgId
}
