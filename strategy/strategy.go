package strategy

import (
	"github.com/anakin/mock/dbops"
	"strings"
	"time"
)

type Strategy struct {
	Id            int
	Name          string
	StrategyType  int
	Amount        int
	Keep          int
	BuyDuration   []string
	BuyDirection  int
	BuyMa         int
	SellDuration  []string
	SellDirection int
	SellMa        int
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
	tableMap = map[string]string{KLINE_TIME_MIN: "kline_min", KLINE_TIME_FIVE_MIN: "kline_five_min", KLINE_TIME_FIF_MIN: "kline_fif_min", KLINE_TIME_THIR_MIN: "kline_thir_min", KLINE_TIME_HOUR: "kline_hour", KLINE_TIME_FOUR_HOUR: "kline_four_hour", KLINE_TIME_SIX_HOUR: "kline_six_hour", KLINE_TIME_TWL_HOUR: "kline_twl_hour", KLINE_TIME_DAY: "kline_day"}
)

func InitStrategy(strategyId int) (*Strategy, error) {
	s, err := dbops.GetStrategy(strategyId)
	if err != nil {
		return nil, err
	}
	return &Strategy{
		Id:            strategyId,
		Name:          s["name"].(string),
		StrategyType:  s["stype"].(int),
		Amount:        s["amount"].(int),
		Keep:          s["keep"].(int),
		BuyDuration:   strings.Split(s["buyDuration"].(string), ","),
		BuyDirection:  s["buyDirection"].(int),
		BuyMa:         s["buyMa"].(int),
		SellDuration:  strings.Split(s["sellDuration"].(string), ","),
		SellDirection: s["sellDirection"].(int),
		SellMa:        s["sellMa"].(int),
	}, nil
}

func (s *Strategy) CanBuy(startTime time.Time) bool {
	for _, du := range s.BuyDuration {
		ma, err := dbops.MARaise(tableMap[du], startTime, s.BuyMa+1)
		if err != nil {
			return false
		} else if s.BuyDirection == 0 && !ma || s.BuyDirection == 1 && ma {
			return false
		}
	}
	return true
}

func (s *Strategy) CanSell(startTime time.Time) bool {
	for _, du := range s.SellDuration {
		ma, err := dbops.MARaise(tableMap[du], startTime, s.SellMa+1)
		if err != nil {
			return false
		} else if s.SellDirection == 0 && !ma || s.SellDirection == 1 && ma {
			return false
		}
	}
	return true
}
