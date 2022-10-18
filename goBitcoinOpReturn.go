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

func (opReturn *OpReturn) convertTextToHex() (err error) {
	opReturn.MessageHex = hex.EncodeToString([]byte(opReturn.Message))
	return
}

func calFee(countTxIns int, countTxOuts int) (fee float64) {
	//	P2PKH
	// 	Overhead	10 	vbytes
	//  Inputs		148	vbytes x countTxIns
	//  Outputs		34	vbytes x countTxOuts
	satFeesPerByte := 5 // TODO : satFeesPerByte
	fee = float64((10+countTxIns*148+countTxOuts*34)*satFeesPerByte) / 100000000
	return
}

func (opReturn *OpReturn) selectUnspentsForSend() (err error) {
	sort.Slice(opReturn.Unspents, func(i, j int) bool {
		return opReturn.Unspents[i].Amount > opReturn.Unspents[j].Amount
	})

	opReturn.Fee = -1.0
	sumAmountTemp := 0.0
	countInUnspents := 0
	for i, unspent := range opReturn.Unspents {
		if unspent.Confirmations < opReturn.Confirmations {
			continue
		}
		countInUnspents += 1
		opReturn.Unspents[i].Expected = true

		sumAmountTemp += unspent.Amount

		tFee := calFee(countInUnspents, 2)
		if sumAmountTemp >= tFee {
			opReturn.Fee = tFee
			break
		}
	}
	if opReturn.Fee <= 0.0 {
		err = fmt.Errorf("@selectUnspentsForSend(): not sufficient: sumUnspentAmount[%f] < fee[%f]", sumAmountTemp, opReturn.Fee)
		return
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

	// 2. 3. Deprecate

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

type Payment struct {
	RpcUser                   string
	RpcPW                     string
	RpcConnect                string
	RpcPort                   string
	RpcPath                   string
	Address                   string
	PrivKey                   string
	PayInfos                  map[string]float64
	Unspents                  []Unspent
	Confirmations             int
	Fee                       float64
	AmountBalanceUsedUnspends float64
	RawTx                     string
	SignedRawTx               string
	PaymentTxID               string
}

func (payment *Payment) Run() (err error) {
	bitcoinCli := goBitcoinCli.BitcoinRpc{
		RpcUser:    payment.RpcUser,
		RpcPW:      payment.RpcPW,
		RpcConnect: payment.RpcConnect,
		RpcPort:    payment.RpcPort,
		RpcPath:    payment.RpcPath,
	}

	// 1. ListUnspent
	listUnspents, err := bitcoinCli.ListUnspentOfAddress(payment.Address)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
		return
	}
	payment.Unspents = make([]Unspent, 0)
	for _, lUnspent := range listUnspents {
		unspent := Unspent{}
		unspent.TxID = lUnspent["txid"].(string)
		unspent.Vout = (int)(lUnspent["vout"].(float64))
		unspent.Amount = lUnspent["amount"].(float64)
		unspent.Confirmations = (int)(lUnspent["confirmations"].(float64))
		unspent.Expected = false
		payment.Unspents = append(payment.Unspents, unspent)
	}
	sort.Slice(payment.Unspents, func(i, j int) bool {
		return payment.Unspents[i].Amount > payment.Unspents[j].Amount
	})
	payment.Confirmations = 3

	// 4. selectUnspentsForSend
	if err = payment.selectUnspentsForSend(); err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
		return
	}

	// 6. CreateRawTransaction
	createTxUnSpents := make([]map[string]interface{}, 0)
	for _, unspent := range payment.Unspents {
		if !unspent.Expected {
			continue
		}
		tCreateTxUnSpent := make(map[string]interface{})
		tCreateTxUnSpent["txid"] = unspent.TxID
		tCreateTxUnSpent["vout"] = unspent.Vout
		createTxUnSpents = append(createTxUnSpents, tCreateTxUnSpent)
	}

	payment.RawTx, err = bitcoinCli.CreateRawTransaction(createTxUnSpents, payment.PayInfos, "")
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
		return
	}

	// 7. DumpPrivateKey
	if payment.PrivKey == "" {
		payment.PrivKey, err = bitcoinCli.DumpPrivateKey(payment.Address)
		if err != nil {
			err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
			return
		}
	}

	// 8. SignRawTransactionWithKey
	payment.SignedRawTx, err = bitcoinCli.SignRawTransactionWithKey(payment.RawTx, payment.PrivKey)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
		return
	}

	// 9. SendRawTransaction
	payment.PaymentTxID, err = bitcoinCli.SendRawTransaction(payment.SignedRawTx)
	if err != nil {
		err = fmt.Errorf("error!!! goBitcoinOpReturn @payment.Run(): %s", err)
		return
	}

	return
}

func (payment *Payment) selectUnspentsForSend() (err error) {

	sort.Slice(payment.Unspents, func(i, j int) bool {
		return payment.Unspents[i].Amount > payment.Unspents[j].Amount
	})

	//
	// Validation amount -1 : only 1 or 0
	totalAmountCaseCount := 0
	totalAmountCaseAddress := ""
	for tAddress, tAmount := range payment.PayInfos {
		if tAmount < 0.0 {
			totalAmountCaseCount += 1
			totalAmountCaseAddress = tAddress
		}
	}
	hasTotalAmountCase := false
	switch totalAmountCaseCount {
	case 0:
		hasTotalAmountCase = false
	case 1:
		hasTotalAmountCase = true
	default: // more than 2
		err = fmt.Errorf("@payment.selectUnspentsForSend(): incorrect payment.PayInfos: totalAmountCases")
		return
	}

	countPayment := len(payment.PayInfos)
	sumPaymentAmount := 0.0
	for _, tAmount := range payment.PayInfos {
		if tAmount >= 0.0 {
			sumPaymentAmount += tAmount
		}
	}

	sumSelectedUnspentsAmount := 0.0
	countSelectedUnspents := 0
	validSelectedUnspents := false
	for i, unspent := range payment.Unspents {

		if unspent.Confirmations < payment.Confirmations {
			payment.Unspents[i].Expected = false
			continue
		}

		payment.Unspents[i].Expected = true
		sumSelectedUnspentsAmount += unspent.Amount
		countSelectedUnspents += 1
		payment.Fee = calFee(countSelectedUnspents, countPayment)
		if sumSelectedUnspentsAmount >= payment.Fee+sumPaymentAmount {
			validSelectedUnspents = true
			if !hasTotalAmountCase {
				break
			}
		}
	}

	if !validSelectedUnspents {
		err = fmt.Errorf("@payment.selectUnspentsForSend(): not sufficient: sumSelectedUnspentsAmount[%f] < fee[%f]+sumPaymentAmount[%f]", sumSelectedUnspentsAmount, payment.Fee, sumPaymentAmount)
		return
	}

	switch hasTotalAmountCase {
	case true:
		payment.AmountBalanceUsedUnspends = 0.0
		tTotalAmount := math.Round((sumSelectedUnspentsAmount-payment.Fee-sumPaymentAmount)*100000000) / 100000000
		if tTotalAmount > 0.0 {
			payment.PayInfos[totalAmountCaseAddress] = tTotalAmount // Update minus-amount-value to final-amount-value[tTotalAmount]
		}
	case false:
		payment.AmountBalanceUsedUnspends = math.Round((sumSelectedUnspentsAmount-payment.Fee-sumPaymentAmount)*100000000) / 100000000
		if payment.AmountBalanceUsedUnspends > 0.0 {
			payment.PayInfos[payment.Address] = payment.AmountBalanceUsedUnspends // Add payment.PayInfos{payment.Address:AmountBalanceUsedUnspends}
		}
	}

	return
}
