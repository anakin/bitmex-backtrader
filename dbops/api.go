package dbops

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"strconv"
	"time"
)

type KData struct {
	Ktime time.Time
	Close float64
}

func GetMockById(mockId int) (map[string]interface{}, error) {
	var (
		id          int
		sId         int
		start       time.Time
		end         time.Time
		leverage    int
		init_base   int64
		cal_rate    int
		cal_funding int
		taker_rate  float64
	)
	smt, err := dbConn.Prepare("select id,strategy_id,start_time,end_time,leverage,taker_rate,init_base,cal_rate,cal_funding from mock where id = ?")
	if err != nil {
		return nil, err
	}
	err1 := smt.QueryRow(mockId).Scan(&id, &sId, &start, &end, &leverage, &taker_rate, &init_base, &cal_rate, &cal_funding)
	if err1 != nil {
		return nil, err1
	}

	out := map[string]interface{}{
		"id":         id,
		"sid":        sId,
		"startDate":  start,
		"endDate":    end,
		"leverage":   leverage,
		"takerRate":  taker_rate,
		"initBase":   init_base,
		"calRate":    cal_rate,
		"calFunding": cal_funding,
	}
	return out, nil
}

func GetMocks() (map[string]interface{}, error) {
	var (
		id        int
		sId       int
		start     time.Time
		end       time.Time
		leverage  int
		init_base int
	)
	smt, err := dbConn.Prepare("select id,strategy_id,start_time,end_time,leverage,init_base from mock where status = 0 limit 1")
	if err != nil {
		return nil, err
	}
	err1 := smt.QueryRow().Scan(&id, &sId, &start, &end, &leverage, &init_base)
	if err1 != nil {
		return nil, err1
	}

	out := map[string]interface{}{
		"id":        id,
		"sid":       sId,
		"startDate": start,
		"endDate":   end,
		"leverage":  leverage,
		"initBase":  init_base,
	}
	return out, nil
}

func GetStrategy(strategyId int) (map[string]interface{}, error) {
	var (
		name       string
		data       string
		buy_kline  string
		sell_kline string
		stype      int
		keep       int
	)

	stmtOut, err := dbConn.Prepare("select name,data,type,buy_kline,sell_kline,keep from kline_strategy where id=?")
	if err != nil {
		log.Println("get strategy error:", err.Error())
		return nil, err
	}
	err1 := stmtOut.QueryRow(strategyId).Scan(&name, &data, &stype, &buy_kline, &sell_kline, &keep)
	if err1 != nil {
		log.Println("strategy query error:", err.Error())
		return nil, err
	}
	_ = stmtOut.Close()
	amount, _ := strconv.ParseInt(gjson.Get(data, "buy_amount").String(), 10, 64)
	buy_direction, _ := strconv.Atoi(gjson.Get(data, "buy_direction").String())

	buy_ma, _ := strconv.Atoi(gjson.Get(data, "buy_ma").String())
	sell_direction, _ := strconv.Atoi(gjson.Get(data, "sell_direction").String())

	sell_ma, _ := strconv.Atoi(gjson.Get(data, "sell_ma").String())

	s := map[string]interface{}{
		"name":          name,
		"stype":         stype,
		"amount":        amount,
		"keep":          keep,
		"buyDuration":   buy_kline,
		"buyDirection":  buy_direction,
		"buyMa":         buy_ma,
		"sellDuration":  sell_kline,
		"sellDirection": sell_direction,
		"sellMa":        sell_ma,
	}
	return s, nil
}

func GetFundingRate(t time.Time) float64 {
	var (
		rate float64
	)
	smt, err := dbConn.Prepare("select FundingRate from funding_rate where Timestamp=?")
	if err != nil {
		log.Println("prepare sql error,", err.Error())
	}
	err = smt.QueryRow(t).Scan(&rate)
	if err != nil {
		log.Println("funding error,", err.Error())
		return float64(0)
	}
	_ = smt.Close()
	return rate
}

func GetKline(table string, startTime time.Time, limit int) ([]KData, error) {
	var (
		out   []KData
		ktime time.Time
		close float64
	)
	sql := fmt.Sprintf("select ktime,close from %s  where ktime <= ? order by ktime desc limit ?", table)
	stmtOut, err := dbConn.Prepare(sql)
	if err != nil {
		log.Println("get kline error:", err.Error())
		return nil, err
	}
	row, err := stmtOut.Query(startTime, limit)
	if err != nil {
		log.Println("query kline error:", err.Error())
		return nil, err
	}
	for row.Next() {
		err1 := row.Scan(&ktime, &close)
		if err1 != nil {
			log.Println("return kline error:", err1.Error())
			break
		}
		out = append(out, KData{
			Ktime: ktime,
			Close: close,
		})
	}
	_ = row.Close()
	return out, nil
}

func MARaise(tableName string, startTime time.Time, limit int) (bool, error) {
	data, err := GetKline(tableName, startTime, limit)
	if err != nil {
		log.Println("get kline error", err.Error())
		return false, err
	}
	if len(data) < limit {
		err1 := errors.New("not enough data")
		return false, err1
	}
	//log.Println("data:", data)
	var total1, total2 float64
	data1 := data[:limit-1]
	data2 := data[1:]
	for _, v1 := range data1 {
		total1 += v1.Close
	}
	for _, v2 := range data2 {
		total2 += v2.Close
	}
	ma1 := total1 / float64(limit-1)
	ma2 := total2 / float64(limit-1)
	//log.Println("ma1:", ma1, ";ma2:", ma2)
	return ma1 > ma2, nil
}

func GetLastPrice(t time.Time) (float64, error) {
	var price float64
	smt, err := dbConn.Prepare("select close from kline_min where ktime < ? order by ktime desc limit 1")
	if err != nil {
		return float64(0), err
	}
	err1 := smt.QueryRow(t).Scan(&price)
	if err != nil {
		return float64(0), err1
	}
	_ = smt.Close()
	return price, nil
}

func AddMockLog(mockId int, ktime time.Time, profit float64, op int, amount int64, price float64, rate float64) error {
	smt, err := dbConn.Prepare("insert into mock_log (mock_id,ktime,profit,op,amount,price,rate) values(?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	_, err1 := smt.Exec(mockId, ktime, profit, op, amount, price, rate)
	if err1 != nil {
		return err1
	}
	_ = smt.Close()
	return nil
}

func UpMock(mockId int, status int, drawdown float64) error {
	smt, err := dbConn.Prepare("update mock set status=?,drawdown=? where id=?")
	if err != nil {
		log.Println("prepare update sql error", err.Error())
		return err
	}
	_, err = smt.Exec(status, drawdown, mockId)
	if err != nil {
		log.Println("exec update sql error", err.Error())
		return err
	}
	//affect, err := res.RowsAffected()
	//log.Println("update affect:", affect)
	_ = smt.Close()
	return nil
}

func ClearLog(mockId int) error {
	smt, err := dbConn.Prepare("delete from mock_log where mock_id=?")
	if err != nil {
		return err
	}
	_, err = smt.Exec(mockId)
	if err != nil {
		return err
	}
	return nil

}
