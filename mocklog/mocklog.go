package mocklog

import (
	"github.com/anakin/mock/dbops"
	"time"
)

type MockLog struct {
	MockId int
	Ktime  time.Time
	Profit float64
	Op     int
	Amount int64
	Price  float64
	Rate   float64
}

func (m *MockLog) Clear() {
	_ = dbops.ClearLog(m.MockId)
}
func (m *MockLog) Add() {
	_ = dbops.AddMockLog(m.MockId, m.Ktime, m.Profit, m.Op, m.Amount, m.Price, m.Rate)
}
