package models

import (
	"encoding/json"
	"fmt"
	"log"
	"stock-with-alpha/config"
	"strings"
	"time"
)

type SignalEvent struct{
	Time time.Time `json:"time"`
	Symbol string `json:"symbol"`
	Side string `json:"side"`
	Price float64 `json:"price"`
	Size float64 `json:"size"`
}

func (s *SignalEvent) Save() bool{
	cmd := fmt.Sprintf("INSERT INTO %s (time, symbol, side, price, size) VALUES(?, ?, ?, ?, ?)", tableNameSignalEvents)
	_, err := DbConnection.Exec(cmd, s.Time, s.Symbol, s.Side, s.Price, s.Size)
	if err != nil{
		if strings.Contains(err.Error(), "UNIQUE constraint failed"){
			log.Println(err)
			// レコードが既にあるのでtrueを返す
			return true
		}
		return false
	}
	return true
}

// 今回は買う→売る→買う...というような交互に取引することを制限する。
//それを管理するのがこのスライスを持つ構造体
type SignalEvents struct{
	Signals []SignalEvent `json:"signals,omitempty"`
}

func NewSignalEvents() *SignalEvents{
	return &SignalEvents{}
}

func GetSignalEventsByCount(loadEvents int) *SignalEvents{
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, symbol, side, price, size FROM %s WHERE symbol = ? ORDER BY time DESC LIMIT ?
	) ORDER BY time ASC;`, tableNameSignalEvents)
	rows, err := DbConnection.Query(cmd, config.Config.Symbol, loadEvents)
	if err != nil{
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents
	for rows.Next(){
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.Symbol, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}
	err = rows.Err()
	if err != nil{
		return nil
	}

	return &signalEvents
}


func GetSignalEventsAfterTime(timeTime time.Time) *SignalEvents{
	cmd := fmt.Sprintf(`SELECT * FROM (
		SELECT time, symbol, side, price, size FROM %s
		WHERE DATETIME(time) >= DATETIME(?)
		ORDER BY time DESC
	) ORDER BY time ASC;`, tableNameSignalEvents)

	rows, err := DbConnection.Query(cmd, timeTime)
	if err != nil{
		return nil
	}
	defer rows.Close()

	var signalEvents SignalEvents

	for rows.Next(){
		var signalEvent SignalEvent
		rows.Scan(&signalEvent.Time, &signalEvent.Symbol, &signalEvent.Side, &signalEvent.Price, &signalEvent.Size)
		signalEvents.Signals = append(signalEvents.Signals, signalEvent)
	}

	return &signalEvents
}


func (s *SignalEvents) CanBuy(time time.Time) bool{
	lenSignals := len(s.Signals)

	if lenSignals == 0{
		return true
	}

	lastSignal := s.Signals[lenSignals - 1]
	if lastSignal.Side == "SELL" && lastSignal.Time.Before(time){
		return true
	}

	return false
}


func(s *SignalEvents) Buy(symbol string, time time.Time, price, size float64, save bool) bool {
	if !s.CanBuy(time){
		return false
	}

	signalEvent := SignalEvent{
		Time: time,
		Symbol: symbol,
		Side: "BUY",
		Size: size,
		Price: price,
	}

	if save{
		signalEvent.Save()
	}
	s.Signals = append(s.Signals, signalEvent)

	return true
}


func (s *SignalEvents) CanSell(time time.Time) bool{
	lenSignals := len(s.Signals)

	if lenSignals == 0{
		return false
	}

	lastSignal := s.Signals[lenSignals - 1]
	if lastSignal.Side == "BUY" && lastSignal.Time.Before(time){
		return true
	}

	return false
}

func (s *SignalEvents) Sell(symbol string, time time.Time, price, size float64, save bool) bool{
	if !s.CanSell(time){
		return false
	}

	signalEvent := SignalEvent{
		Symbol: symbol,
		Time: time,
		Side: "SELL",
		Size: size,
		Price: price,
	}

	if save{
		signalEvent.Save()
	}

	s.Signals = append(s.Signals, signalEvent)

	return true
}


func (s *SignalEvents) Profit() float64{
	total := 0.0
	beforeSell := 0.0
	isHolding := false
	for i, signalEvent := range s.Signals{
		if i == 0 && signalEvent.Side == "SELL"{
			continue
		}
		if signalEvent.Side == "BUY"{
			total -= signalEvent.Price * signalEvent.Size
			isHolding = true
		}
		if signalEvent.Side == "SELL"{
			total += signalEvent.Price * signalEvent.Size
			isHolding = false
			beforeSell = total
		}
	}
	if isHolding{
		return beforeSell
	}

	return total
}

// Sell, Buy関数によって得られたプロフィットをsignalEevnts構造体に新たなフィールドとして加える
func (s SignalEvents) MarshalJSON() ([] byte, error){
	value, err := json.Marshal(&struct{
		Signals []SignalEvent `json:"signals,omitempty"`
		Profit float64 `json:"profit,omitempty"`
	}{
		// プロフィットの計算はオブジェクトのエンコード時に行う
		Signals: s.Signals,
		Profit: s.Profit(),
	})
	if err != nil{
		return nil, err
	}
	return value, err
}


func (s *SignalEvents) CollectAfter(time time.Time) *SignalEvents{
	for i, signal := range s.Signals{

		if time.After(signal.Time){
			continue
		}
		return &SignalEvents{ Signals: s.Signals[i:] }
	}
	return nil
}
