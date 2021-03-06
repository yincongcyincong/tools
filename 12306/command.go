package main

import (
	"errors"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/tools/12306/conf"
	"github.com/tools/12306/module"
	"github.com/tools/12306/notice"
	"github.com/tools/12306/utils"
	"math"
	"strings"
	"time"
)

func CommandStart() {
	var err error
	if err = GetLoginData(); err != nil {
		qrImage, err := CreateImage()
		if err != nil {
			seelog.Errorf("创建二维码失败:%v", err)
			return
		}
		qrImage.Image = ""

		err = QrLogin(qrImage)
		if err != nil {
			seelog.Errorf("登陆失败:%v", err)
			return
		}
	}
	searchParam := new(module.SearchParam)
	var trainStr, seatStr, passengerStr string
	for i := 1; i < math.MaxInt64; i++ {
		getUserInfo(searchParam, &trainStr, &seatStr, &passengerStr)
		if trainStr != "" && seatStr != "" && passengerStr != "" {
			break
		}

		time.Sleep(5 * time.Second)
	}

	// 开始轮训买票
	trainMap := utils.GetBoolMap(strings.Split(trainStr, "#"))
	passengerMap := utils.GetBoolMap(strings.Split(passengerStr, "#"))
	seatSlice := strings.Split(seatStr, "#")

Search:
	var trainData *module.TrainData
	for i := 0; i < math.MaxInt64; i++ {
		trainData, err = getTrainInfo(searchParam, i, trainMap, seatSlice)
		if err == nil {
			break
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	seelog.Info("开始购买", trainData.TrainNo)
	err = startOrder(searchParam, trainData, passengerMap)
	if err != nil {
		utils.AddBlackList(trainData.TrainNo)
		goto Search
	}

	// 购买完成后自动退出登陆，避免出现多次登陆的情况
	GetLoginData()
	LoginOut()

	notice.SendWxrootMessage(*wxrobot, fmt.Sprintf("车次：%s 购买成功, 请登陆12306查看", trainData.TrainNo))

}

func getTrainInfo(searchParam *module.SearchParam, i int, trainMap map[string]bool, seatSlice []string) (*module.TrainData, error) {

	var err error
	searchParam.SeatType = ""
	var trainData *module.TrainData

	// 一分钟进行一次自动登陆
	if i%30 == 0 {
		err = GetLoginData()
		if err != nil {
			seelog.Errorf("自动登陆失败：%v", err)
		}
	}

	trains, err := GetTrainInfo(searchParam)
	if err != nil {
		seelog.Errorf("查询车站失败:%v", err)
		return nil, err
	}

	for _, t := range trains {
		// 在选中的，但是不在小黑屋里面
		if utils.InBlackList(t.TrainNo) {
			seelog.Info(t.TrainNo, "在小黑屋，需等待60s")
			continue
		}

		if trainMap[t.TrainNo] {
			for _, s := range seatSlice {
				if t.SeatInfo[s] != "" && t.SeatInfo[s] != "无" {
					trainData = t
					searchParam.SeatType = conf.OrderSeatType[s]
					break
				}
				seelog.Infof("%s %s 数量: %s", t.TrainNo, s, t.SeatInfo[s])
			}

			if searchParam.SeatType != "" {
				break
			}
		}
	}

	if trainData == nil || searchParam.SeatType == "" {
		seelog.Info("暂无车票可以购买")
		return nil, errors.New("暂无车票可以购买")
	}
	return trainData, nil
}

func getUserInfo(searchParam *module.SearchParam, trainStr, seatStr, passengerStr *string) {
	fmt.Println("请输入日期 起始站 到达站: ")
	fmt.Scanf("%s %s %s", &searchParam.TrainDate, &searchParam.FromStationName, &searchParam.ToStationName)
	searchParam.FromStation = conf.Station[searchParam.FromStationName]
	searchParam.ToStation = conf.Station[searchParam.ToStationName]

	trains, err := GetTrainInfo(searchParam)
	if err != nil {
		seelog.Errorf("查询车站失败:%v", err)
		return
	}
	for _, t := range trains {
		fmt.Println(fmt.Sprintf("车次: %s, 状态: %s, 始发车站: %s, 终点站:%s,  %s: %s, 历时：%s, 二等座: %s, 一等座: %s, 商务座: %s, 软卧: %s, 硬卧: %s，软座: %s，硬座: %s， 无座: %s,",
			t.TrainNo, t.Status, t.FromStationName, t.ToStationName, t.StartTime, t.ArrivalTime, t.DistanceTime, t.SeatInfo["二等座"], t.SeatInfo["一等座"], t.SeatInfo["商务座"], t.SeatInfo["软卧"], t.SeatInfo["硬卧"], t.SeatInfo["软座"], t.SeatInfo["硬座"], t.SeatInfo["无座"]))
	}

	fmt.Println("请输入车次(多个#分隔):")
	fmt.Scanf("%s", trainStr)

	fmt.Println("请输入座位类型(多个#分隔，一等座，二等座，硬座，软卧，硬卧等):")
	fmt.Scanf("%s", seatStr)

	submitToken, err := GetRepeatSubmitToken()
	if err != nil {
		seelog.Errorf("获取提交数据失败:%v", err)
		return
	}
	passengers, err := GetPassengers(submitToken)
	if err != nil {
		seelog.Errorf("获取用户失败:%v", err)
		return
	}
	for _, p := range passengers.Data.NormalPassengers {
		fmt.Println(fmt.Sprintf("乘客姓名：%s", p.PassengerName))
	}

	fmt.Println("请输入乘客姓名(多个#分隔): ")
	fmt.Scanf("%s", passengerStr)

	return
}

func startOrder(searchParam *module.SearchParam, trainData *module.TrainData, passengerMap map[string]bool) error {
	err := GetLoginData()
	if err != nil {
		seelog.Errorf("自动登陆失败：%v", err)
	}

	err = CheckUser()
	if err != nil {
		seelog.Errorf("检查用户状态失败：%v", err)
		return err
	}

	err = SubmitOrder(trainData, searchParam)
	if err != nil {
		seelog.Errorf("提交订单失败：%v", err)
		return err
	}

	submitToken, err := GetRepeatSubmitToken()
	if err != nil {
		seelog.Errorf("获取提交数据失败：%v", err)
		return err
	}

	passengers, err := GetPassengers(submitToken)
	if err != nil {
		seelog.Errorf("获取乘客失败：%v", err)
		return err
	}
	buyPassengers := make([]*module.Passenger, 0)
	for _, p := range passengers.Data.NormalPassengers {
		if passengerMap[p.PassengerName] {
			buyPassengers = append(buyPassengers, p)
		}
	}

	err = CheckOrder(buyPassengers, submitToken, searchParam)
	if err != nil {
		seelog.Errorf("检查订单失败：%v", err)
		return err
	}

	err = GetQueueCount(submitToken, searchParam)
	if err != nil {
		seelog.Errorf("获取排队数失败：%v", err)
		return err
	}

	err = ConfirmQueue(buyPassengers, submitToken, searchParam)
	if err != nil {
		seelog.Errorf("提交订单失败：%v", err)
		return err
	}

	var orderWaitRes *module.OrderWaitRes
	for i := 0; i < 12; i++ {
		orderWaitRes, err = OrderWait(submitToken)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if orderWaitRes.Data.OrderId != "" {
			break
		}
	}

	if orderWaitRes != nil {
		err = OrderResult(submitToken, orderWaitRes.Data.OrderId)
		if err != nil {
			seelog.Errorf("获取订单状态失败：%v", err)
		}
	}

	if orderWaitRes == nil || orderWaitRes.Data.OrderId == "" {
		seelog.Infof("购买成功")
		return nil
	}

	seelog.Infof("购买成功，订单号：%s", orderWaitRes.Data.OrderId)
	return nil
}
