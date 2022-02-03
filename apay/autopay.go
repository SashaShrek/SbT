package apay

import (
	"fmt"

	"github.com/SashaShrek/db"
)

func (d Data) GetAutoPay(id int) (bool, int) {
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
