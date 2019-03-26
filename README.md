# bitmex-backtrader
Golang

这是一个Golang开发的Bitmex交易所的策略回测系统


## 使用方法：

目前支持的策略：根据历史k线的MA（数量可以定义）和最新价格来确定买卖信号

首先在mysql数据库中保存策略，具体的字段类型和策略定义：

	        Id:           策略ID,
		Name:          策略名称
		StrategyType:  策略类型（预留，支持多种策略类型）
		Amount:        买卖的数量
		Keep:          保留字段
		BuyDuration:   判断买入的k线值（例如，5min,1hour）
		BuyDirection:  判断买入的方向（涨、跌）
		BuyMa:         买入的ma数量（ma5,ma7）
		SellDuration:  卖出的k线值
		SellDirection: 卖出的方向
		SellMa:        卖出的ma数量

build之后，在后台运行

前台通过api调用后台服务运行

运行完成后，结果数据直接写入数据库

前端直接从数据库获取结果，显示

## api：
/mock/:id

## TODO

支持更多策略类型

支持更多交易所
