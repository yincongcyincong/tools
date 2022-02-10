package main

import (
	"errors"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func SubmitOrder(trainData *module.TrainData, searchParam *module.SearchParam) error {
	var err error
	data := make(url.Values)
	data.Set("train_date", searchParam.TrainDate)
	data.Set("back_train_date", time.Now().Format("2006-01-02"))
	data.Set("tour_flag", "dc")
	data.Set("purpose_codes", "ADULT")
	data.Set("query_from_station_name", "北京")
	data.Set("query_to_station_name", "天津")
	secretStr, err := url.QueryUnescape(trainData.SecretStr)
	if err != nil {
		seelog.Error(err)
		return err
	}
	data.Set("secretStr", secretStr)

	submitOrder := new(module.SubmitOrderRes)
	err = utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/leftTicket/submitOrderRequest", submitOrder, map[string]string{"Referer": "https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc"})
	if err != nil {
		seelog.Error(err)
		return err
	}

	if submitOrder.Data != "0" {
		seelog.Errorf("submit order fail: %+v", submitOrder)
		return errors.New("submit order fail")
	}

	return nil
}

func CheckOrder(passenger []*module.Passenger, submitToken *module.SubmitToken, searchParam *module.SearchParam) error {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型
	data := fmt.Sprintf("bed_level_order_num=000000000000000000000000000000&passengerTicketStr=%s,%s&oldPassengerStr=%s&tour_flag=dc&randCode=&sessionId=&sig=&cancel_flag=2&_json_att=&whatsSelecte=1&scene=nc_login&REPEAT_SUBMIT_TOKEN=%s",
		searchParam.SeatType, strings.Replace(url.QueryEscape(passenger[0].PassengerTicketStr), "%2A", "*", -1), strings.Replace(url.QueryEscape(passenger[0].OldPassengerStr), "%2A", "*", -1), submitToken.Token)

	checkOrderRes := new(module.CheckOrderRes)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo", checkOrderRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		seelog.Error(err)
		return err
	}

	if !checkOrderRes.Status || !checkOrderRes.Data.SubmitStatus {
		seelog.Errorf("check order fail: %+v", checkOrderRes)
		return errors.New("check order fail")
	}

	return nil
}

func OrderResult(submitToken *module.SubmitToken, orderNo string) error {

	// url encode需要小心，会多处理
	var err error
	data := make(url.Values)
	data.Set("orderSequence_no", orderNo)
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken.Token)
	data.Set("json_att", "")

	orderRes := new(module.OrderResultRes)
	err = utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/resultOrderForDcQueue", orderRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		seelog.Error(err)
		return err
	}

	if !orderRes.Data.SubmitStatus {
		seelog.Errorf("result order fail: %+v", orderRes)
		return err
	}

	return nil
}

func OrderWait(submitToken *module.SubmitToken) (*module.OrderWaitRes, error) {

	// url encode需要小心，会多处理
	var err error
	orderWaitUrl := fmt.Sprintf("https://kyfw.12306.cn/otn/confirmPassenger/queryOrderWaitTime?random=%s&tourFlag=dc&_json_att=&REPEAT_SUBMIT_TOKEN=%s", "16442323111232", submitToken.Token)
	orderWaitRes := new(module.OrderWaitRes)
	err = utils.RequestGet(utils.GetCookieStr(), orderWaitUrl, orderWaitRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		seelog.Error(err)
		return nil, err
	}

	if orderWaitRes.Data.OrderId != "" {
		return orderWaitRes, nil
	} else {
		switch orderWaitRes.Data.WaitTime {
		case -100:
			seelog.Info("重新获取订单号")
		case -2, -3:
			seelog.Errorf("订单失败获取消")
		default:
			seelog.Infof("等待时间:%d,等待人数：%d", orderWaitRes.Data.WaitTime, orderWaitRes.Data.WaitCount)
		}
		return nil, errors.New("需要继续等待")
	}
}

func ConfirmQueue(passenger []*module.Passenger, submitToken *module.SubmitToken, searchParam *module.SearchParam) error {

	// url encode需要小心，会多处理
	data := fmt.Sprintf("passengerTicketStr=%s,%s&oldPassengerStr=%s&purpose_codes=%s&key_check_isChange=%s&leftTicketStr=%s&train_location=%s&seatDetailType=000&roomType=00&whatsSelecte=1&dwAll=N&_json_at=&randCode=&choose_seats=1D&REPEAT_SUBMIT_TOKEN=%s&is_jy=N&is_cj=Y&encryptedData=%s",
		searchParam.SeatType, strings.Replace(url.QueryEscape(passenger[0].PassengerTicketStr), "%2A", "*", -1), strings.Replace(url.QueryEscape(passenger[0].OldPassengerStr), "%2A", "*", -1),
		submitToken.TicketInfo["purpose_codes"].(string), submitToken.TicketInfo["key_check_isChange"].(string), submitToken.TicketInfo["leftTicketStr"].(string), submitToken.TicketInfo["train_location"].(string), submitToken.Token)

	confirmQueue := new(module.ConfirmQueueRes)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/confirmSingleForQueue", confirmQueue, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		seelog.Error(err)
		return err
	}

	switch data := confirmQueue.Data.(type) {
	case string:
		seelog.Error(data)
		return errors.New(data)
	case module.ConfirmData:
		if !data.SubmitStatus {
			seelog.Errorf("confirm queue fail: %+v",confirmQueue.Data)
			return errors.New("confirm queue fail")
		}
	}

	return nil
}

func GetQueueCount(submitToken *module.SubmitToken, searchParam *module.SearchParam) error {
	var err error
	startTime, err := time.Parse("2006-01-02", searchParam.TrainDate)
	if err != nil {
		seelog.Error(err)
		return err
	}

	data := fmt.Sprintf("train_location=%s&purpose_codes=%s&_json_att=&leftTicket=%s&toStationTelecode=%s&fromStationTelecode=%s&REPEAT_SUBMIT_TOKEN=%s&seatType=O&train_no=%s&stationTrainCode=%s&train_date=%s",
		submitToken.TicketInfo["train_location"].(string), submitToken.TicketInfo["purpose_codes"].(string), submitToken.TicketInfo["leftTicketStr"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["to_station"].(string), submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["from_station"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["from_station"].(string), submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["train_no"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["station_train_code"].(string),
		strings.Replace(
			strings.Replace(
				url.QueryEscape(fmt.Sprintf("%s 00:00:00 GMT+0800 (中国标准时间)", startTime.Format("Mon Jan 02 2006"))), "%28", "(", -1),
			"%29", ")", -1),
	)

	queueRes := new(module.QueueCountRes)
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/getQueueCount", queueRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		seelog.Error(err)
		return err
	}

	if !queueRes.Status {
		// todo 开启小黑屋
		return errors.New("购买失败，开启小黑屋")
	}

	ticketNum, _ := strconv.Atoi(queueRes.Data.Ticket)
	if queueRes.Data.Ticket != "充足" && ticketNum <= 0 {
		seelog.Warn("开始购买无座")
		return nil
	}

	if queueRes.Data.Op2 == "true" {
		seelog.Error(err)
		return errors.New("排队人数超过票数")
	}

	return nil
}

func AutoBuy(passenger *module.Passenger, trainData *module.TrainData, submitToken *module.SubmitToken) {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型
	var err error

	data := make(url.Values)
	data.Set("bed_level_order_num", "000000000000000000000000000000")
	data.Set("passengerTicketStr", "O,"+passenger.PassengerTicketStr)
	data.Set("oldPassengerStr", passenger.OldPassengerStr)
	data.Set("tour_flag", "dc")
	data.Set("cancel_flag", "2")
	data.Set("purpose_codes", "ADULT")
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken.Token)
	data.Set("query_from_station_name", trainData.FromStation)
	data.Set("query_to_station_name", trainData.ToStation)
	data.Set("train_date", trainData.StartTime)

	trainData.SecretStr, err = url.QueryUnescape(trainData.SecretStr)
	if err != nil {
		log.Panicln(err)
	}
	data.Set("secretStr", trainData.SecretStr)

	qrImage := new(module.AutoBuyRes)
	err = utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/autoSubmitOrderRequest", qrImage, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		log.Panicln(err)
	}
}
