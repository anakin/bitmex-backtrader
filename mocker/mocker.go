package mocker

import (
	"github.com/anakin/mock/dbops"
	"github.com/anakin/mock/mocklog"
	"github.com/anakin/mock/strategy"
	"log"
	"math"
	"time"
)

type Mock struct {
	Id         int
	StrategyId int
	InitBase   int64
	StartDate  time.Time
	EndDate    time.Time
	TakerRate  float64
	CalRate    int
	CalFunding int
	Leverage   int
	S          *strategy.Strategy
}

const (
	KLINE_TIME_MIN       = "1m"
	KLINE_TIME_FIVE_MIN  = "5m"
	KLINE_TIME_FIF_MIN   = "15m"
	KLINE_TIME_THIR_MIN  = "30m"
	KLINE_TIME_HOUR      = "1h"
	KLINE_TIME_FOUR_HOUR = "4h"
	KLINE_TIME_SIX_HOUR  = "6h"
	KLINE_TIME_TWL_HOUR  = "12h"
	KLINE_TIME_DAY       = "1d"
)

var (
	base       float64
	profit     = float64(0)
	position   = int64(0)
	max_dd     = float64(-9999)
	endTime, _ = time.Parse("2006-01-02 15:04:05", "0001-01-01 00:00:00")
	avgPrice   = float64(0)
)

func InitMocker(m map[string]interface{}) (*Mock, error) {
	sid := m["sid"].(int)
	s, err := strategy.InitStrategy(sid)
	if err != nil {
		return nil, err
	}
	return &Mock{
		Id:         m["id"].(int),
		StrategyId: m["sid"].(int),
		StartDate:  m["startDate"].(time.Time),
		EndDate:    m["endDate"].(time.Time),
		TakerRate:  m["takerRate"].(float64),
		Leverage:   m["leverage"].(int),
		InitBase:   m["initBase"].(int64),
		CalRate:    m["calRate"].(int),
		CalFunding: m["calFunding"].(int),
		S:          s,
	}, nil
}

func (m *Mock) Loop() {
	log.Println("start mock strategy:", m.S.Name)
	log.Println("start time:", m.StartDate)
	log.Println("end time:", m.EndDate)
	log.Println("calculate taker rate:", m.CalRate)
	log.Println("calculate funding rate:", m.CalFunding)
	base = float64(m.InitBase)
	position = 0
	max_dd = float64(-9999)
	avgPrice = float64(0)
	var mockTime time.Time
	t := make(chan time.Time)
	ml := &mocklog.MockLog{
		MockId: m.Id,
	}
	ml.Clear()
	go m.Run(t, ml)
	mockTime = m.StartDate
	for {
		mockTime = mockTime.Add(time.Duration(m.MinTicker()) * time.Minute)
		//log.Println("send time:", mockTime)
		t <- mockTime
		if mockTime == m.EndDate {
			break
		}
	}
	m.UpStatus(2, max_dd)
	t <- endTime
}

func (m *Mock) Run(t chan time.Time, ml *mocklog.MockLog) {
	max_profit := float64(-9999)
	for {
		select {
		case mockTime := <-t:
			if mockTime == endTime {
				t <- mockTime
			} else {
				price, err := dbops.GetLastPrice(mockTime)
				if err != nil {
					log.Println("get current price error,", err.Error())
					return
				} else if price == float64(0) {
					log.Println("price is zero,", mockTime)
					continue
				}
				ml.Ktime = mockTime
				ml.Op = 0
				ml.Amount = 0
				ml.Price = float64(0)
				ml.Rate = float64(0)
				if isTradeTime(m.S.BuyDuration[0], mockTime) {
					if m.S.CanBuy(mockTime) && base > 0 {
						//如果设置了持续加仓
						if m.S.Keep == 1 {
							ml.Op = 1
							ml.Price = price
							ml.Amount = m.Buy(price)
						} else {
							//空仓的时候买入
							if position == 0 {
								ml.Op = 1
								ml.Price = price
								ml.Amount = m.Buy(price)
							}
						}

					}
				}
				if isTradeTime(m.S.SellDuration[0], mockTime) {
					if m.S.CanSell(mockTime) && position > 0 {
						ml.Amount = position
						ml.Op = 2
						ml.Price = price
						m.Sell(price)
					}
				}
				//calculate funding rate here
				if m.CalFunding == 1 && isFundingTime(mockTime) && position > 0 {
					rate := dbops.GetFundingRate(mockTime)
					log.Println("funding rate is:", rate)
					funding := float64(position) / price * (-rate)
					base += funding
				}
				//calculate profit here

				if ml.Op > 0 {
					//v := float64(0)
					//if avgPrice > 0 && position > 0 {
					//	v = float64(position) * (1/avgPrice - 1/price)
					//}
					//profit = (base + v) / init_base
					profit = base / float64(m.InitBase)
					//log.Println(mockTime, ";profit=", profit, ";base:", base, "posi:", float64(position)/price)
					max_profit = math.Max(profit, max_profit)
					max_dd = math.Max(max_dd, max_profit/profit-1)
					ml.Profit = profit
					ml.Rate = 1/avgPrice - 1/price
					go ml.Add()
				}
			}
			break
		}
	}
}

func isFundingTime(t time.Time) bool {
	return t.Minute() == 0 && (t.Hour() == 4 || t.Hour() == 12 || t.Hour() == 20)
}

func isTradeTime(klineTime string, t time.Time) bool {
	switch klineTime {
	case KLINE_TIME_MIN:
		return true
	case KLINE_TIME_FIVE_MIN:
		if t.Minute()%5 == 0 {
			return true
		}
	case KLINE_TIME_FIF_MIN:
		if t.Minute()%15 == 0 {
			return true
		}
	case KLINE_TIME_THIR_MIN:
		if t.Minute()%30 == 0 {
			return true
		}
	case KLINE_TIME_HOUR:
		if t.Minute() == 0 {
			return true
		}
	case KLINE_TIME_FOUR_HOUR:
		if t.Minute() == 0 && t.Hour()%4 == 0 {
			return true
		}
	case KLINE_TIME_SIX_HOUR:
		if t.Minute() == 0 && t.Hour()%6 == 0 {
			return true
		}
	case KLINE_TIME_TWL_HOUR:
		if t.Minute() == 0 && t.Hour()%12 == 0 {
			return true
		}
	case KLINE_TIME_DAY:
		if t.Hour() == 0 && t.Minute() == 0 {
			return true
		}
	}
	return false
}

func (m *Mock) MinTicker() int {
	t := m.S.BuyDuration
	t = append(t, m.S.SellDuration...)
	t1 := StrToInt(t)
	min := t1[0]
	for _, v := range t1 {
		if v < min {
			min = v
		}
	}
	return min
}

func StrToInt(d []string) []int {
	res := []int{}
	for _, v := range d {
		switch v {
		case KLINE_TIME_MIN:
			res = append(res, 1)
		case KLINE_TIME_FIVE_MIN:
			res = append(res, 5)
		case KLINE_TIME_FIF_MIN:
			res = append(res, 15)
		case KLINE_TIME_THIR_MIN:
			res = append(res, 30)
		case KLINE_TIME_HOUR:
			res = append(res, 60)
		case KLINE_TIME_FOUR_HOUR:
			res = append(res, 240)
		case KLINE_TIME_SIX_HOUR:
			res = append(res, 360)
		case KLINE_TIME_TWL_HOUR:
			res = append(res, 720)
		case KLINE_TIME_DAY:
			res = append(res, 1440)
		}
	}
	return res
}

func (m *Mock) Buy(price float64) int64 {
	amount := int64(float64(m.InitBase*m.S.Amount/100) * price)
	//base -= v
	//log.Println("buy base before:", base)
	if m.CalRate == 1 {
		v := float64(amount) / price
		base -= v * m.TakerRate
		//log.Println("buy rate :", m.TakerRate, ";v:", v)
		//log.Println("do buy rate her:", base)
	}
	avgPrice = (avgPrice*float64(position) + float64(amount)*price) / float64(position+amount)
	position += amount
	log.Println("buy amount:", amount, ";price:", price, ";base:", base)
	return amount
}

func (m *Mock) Sell(price float64) {
	if m.CalRate == 1 {
		v := float64(position) / price
		base -= v * m.TakerRate
		//log.Println("do sell rate her")
	}
	rate := 1/avgPrice - 1/price
	base += float64(position) * rate
	log.Println("sell amount", position, ";price:", price, ";base:", base)
	position = 0
}

func (m *Mock) UpStatus(status int, drawdown float64) {
	_ = dbops.UpMock(m.Id, status, drawdown)
}
