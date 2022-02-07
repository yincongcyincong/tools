package main

import (
	"fmt"
	"log"
	"net/url"
)

func Buy(passengerRes *PassengerRes) {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型

	passengerTicketStr := fmt.Sprintf("0,%s,%s,%s,%s,%s,N,%s",
		passengerRes.Data.NormalPassengers[0].PassengerType, passengerRes.Data.NormalPassengers[0].PassengerName,
		passengerRes.Data.NormalPassengers[0].PassengerIdTypeCode, passengerRes.Data.NormalPassengers[0].PassengerIdNo,
		passengerRes.Data.NormalPassengers[0].MobileNo, passengerRes.Data.NormalPassengers[0].AllEncStr)
	oldPassengerStr := fmt.Sprintf("%s,%s,%s,%s_",
		passengerRes.Data.NormalPassengers[0].PassengerName,
		passengerRes.Data.NormalPassengers[0].PassengerIdTypeCode, passengerRes.Data.NormalPassengers[0].PassengerIdNo,
		passengerRes.Data.NormalPassengers[0].PassengerType)

	data := make(url.Values)
	data.Set("bed_level_order_num", "000000000000000000000000000000")
	data.Set("passengerTicketStr", "O,"+passengerTicketStr)
	data.Set("oldPassengerStr", oldPassengerStr)
	data.Set("tour_flag", "dc")
	data.Set("randCode", "")
	data.Set("cancel_flag", "2")
	data.Set("_json_att", "")
	data.Set("REPEAT_SUBMIT_TOKEN", passengerRes.SubmitToken)
	fmt.Println(data)

	qrImage := new(QrImage)
	err := request(data, cookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo", qrImage)
	if err != nil {
		log.Panicln(err)
	}

}

func Confirm(passengerRes *PassengerRes) {

}