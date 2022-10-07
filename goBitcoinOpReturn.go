package gobitcoinopreturn

import (
	"encoding/hex"
	"fmt"
	"math"
	"sort"

	goBitcoinCli "github.com/ideajoo/go-bitcoin-cli-light"
)

type OpReturn struct {
	RpcUser                   string
	RpcPW                     string
	RpcConnect                string
	RpcPort                   string
	RpcPath                   string
	Address                   string
	PrivKey                   string
	Message                   string
	MessageHex                string
	Unspents                  []Unspent
	Confirmations             int
	AmountUnspends            float64
	Fee                       float64
	AmountBalanceUsedUnspends float64
	RawTx                     string
	SignedRawTx               string
	OpRetrunTxID              string
}

type Unspent struct {
	TxID          string
	Vout          int
	Amount        float64
	Confirmations int
	Expected      bool
}

type JsonRpc struct {
	JsonRpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func (opReturn *OpReturn) calAmountUnspents() (err error) {
	balance := 0.0
	for _, unspent := range opReturn.Unspents {
		if unspent.Confirmations < opReturn.Confirmations {
			continue
		}
		balance += unspent.Amount
	}
	opReturn.AmountUnspends = math.Round(balance*100000000) / 100000000
	return
}

func (opReturn *OpReturn) convertTextToHex() (err error) {
	opReturn.MessageHex = hex.EncodeToString([]byte(opReturn.Message))
	return
}

func (opReturn *OpReturn) calFee() (err error) {

	// TODO: Cal fee by bytes
	fee := 0.00005000 // Temp Fee

	opReturn.Fee = fee
	tBalance := math.Round((opReturn.AmountUnspends-fee)*100000000) / 100000000
	if tBalance < 0.0 {
		err = fmt.Errorf("@calFee(): not sufficient: totalUnspentAmount[%f] < fee[%f]", opReturn.AmountUnspends, opReturn.Fee)
		return
	}
	return
}

func (opReturn *OpReturn) selectUnspentsForSend() (err error) {
	sort.Slice(opReturn.Unspents, func(i, j int) bool {
		return opReturn.Unspents[i].Amount > opReturn.Unspents[j].Amount
	})

	sumAmountTemp := 0.0
	for i, unspent := range opReturn.Unspents {
		if unspent.Confirmations < opReturn.Confirmations {
			continue
		}
		if unspent.Amount > 0.0 && sumAmountTemp >= opReturn.Fee {
			continue
		} // for handle case : unspent.Amount == 0.0
		opReturn.Unspents[i].Expected = true
		sumAmountTemp += unspent.Amount
		sumAmountTemp = math.Round((sumAmountTemp)*100000000) / 100000000
	}
	opReturn.AmountBalanceUsedUnspends = math.Round((sumAmountTemp-opReturn.Fee)*100000000) / 100000000
	return
}

func (opReturn *OpReturn) Run() (err error) {
	bitcoinCli := goBitcoinCli.BitcoinRpc{
		RpcUser:    opReturn.RpcUser,
		RpcPW:      opReturn.RpcPW,
		RpcConnect: opReturn.RpcConnect,
		RpcPort:    opReturn.RpcPort,
		RpcPath:    opReturn.RpcPath,
	}

	// 1. ListUnspent
	listUnspents, err := bitcoinCli.ListUnspentOfAddress(opReturn.Address)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}
	opReturn.Unspents = make([]Unspent, 0)
	for _, lUnspent := range listUnspents {
		unspent := Unspent{}
		unspent.TxID = lUnspent["txid"].(string)
		unspent.Vout = (int)(lUnspent["vout"].(float64))
		unspent.Amount = lUnspent["amount"].(float64)
		unspent.Confirmations = (int)(lUnspent["confirmations"].(float64))
		unspent.Expected = false
		opReturn.Unspents = append(opReturn.Unspents, unspent)
	}
	sort.Slice(opReturn.Unspents, func(i, j int) bool {
		return opReturn.Unspents[i].Amount > opReturn.Unspents[j].Amount
	})
	opReturn.Confirmations = 3

	// 2. calAmountUnspents
	if err = opReturn.calAmountUnspents(); err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 3. calFee
	if err = opReturn.calFee(); err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 4. selectUnspentsForSend
	if err = opReturn.selectUnspentsForSend(); err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 5. convertTextToHex
	if err = opReturn.convertTextToHex(); err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 6. CreateRawTransaction
	createTxUnSpents := make([]map[string]interface{}, 0)
	for _, unspent := range opReturn.Unspents {
		if !unspent.Expected {
			continue
		}
		tCreateTxUnSpent := make(map[string]interface{})
		tCreateTxUnSpent["txid"] = unspent.TxID
		tCreateTxUnSpent["vout"] = unspent.Vout
		createTxUnSpents = append(createTxUnSpents, tCreateTxUnSpent)
	}

	opReturn.RawTx, err = bitcoinCli.CreateRawTransaction(createTxUnSpents, map[string]float64{opReturn.Address: opReturn.AmountBalanceUsedUnspends}, opReturn.MessageHex)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 7. DumpPrivateKey
	if opReturn.PrivKey == "" {
		opReturn.PrivKey, err = bitcoinCli.DumpPrivateKey(opReturn.Address)
		if err != nil {
			err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
			return
		}
	}

	// 8. SignRawTransactionWithKey
	opReturn.SignedRawTx, err = bitcoinCli.SignRawTransactionWithKey(opReturn.RawTx, opReturn.PrivKey)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	// 9. SendRawTransaction
	opReturn.OpRetrunTxID, err = bitcoinCli.SendRawTransaction(opReturn.SignedRawTx)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @Run(): %s", err)
		return
	}

	return
}
