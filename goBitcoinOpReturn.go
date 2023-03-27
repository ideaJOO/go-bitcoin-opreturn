package gobitcoinopreturn

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

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
	PayInfos                  map[string]float64
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

func ConvertTextToHex(text string) (hexStr string) {
	hexStr = hex.EncodeToString([]byte(text))
	return
}

func ConvertHexToText(hexStr string) (readableStr string, err error) {
	source := make([]byte, hex.DecodedLen(len(hexStr)))
	_, err = hex.Decode(source, []byte(hexStr))
	if err != nil {
		err = fmt.Errorf("@hex.Decode(): %v", err)
		return
	}
	readableStr = string(source[:])
	return
}

func calFee(countTxIns int, countTxOuts int, feePerVByte float64, address ...string) (fee float64) {

	tAddressType := "P2PKH" // Legacy
	if len(address) > 0 {
		if len(address[0]) > 4 {
			if address[0][0:4] == "bc1q" {
				tAddressType = "P2WPKH"
			}
		}
	}
	switch tAddressType {
	case "P2WPKH":
		//	P2WPKH
		// 	Overhead	10.5 	vbytes
		//  Inputs		68	vbytes x countTxIns
		//  Outputs		31	vbytes x countTxOuts
		fee = math.Ceil((10.5+float64(countTxIns*68+countTxOuts*31))*feePerVByte) / 100000000.0
	default:
		//	P2PKH (Legacy)
		// 	Overhead	10 	vbytes
		//  Inputs		148	vbytes x countTxIns
		//  Outputs		34	vbytes x countTxOuts
		fee = math.Ceil((10.0+float64(countTxIns*148+countTxOuts*34))*feePerVByte) / 100000000.0
	}

	return
}

func (opReturn *OpReturn) selectUnspentsForSend() (err error) {
	sort.Slice(opReturn.Unspents, func(i, j int) bool {
		return opReturn.Unspents[i].Amount > opReturn.Unspents[j].Amount
	})

	payValueExtra := 0.0
	countExtra := 0
	for _, feeVal := range opReturn.PayInfos {
		payValueExtra += feeVal
		countExtra += 1
	}

	opReturn.Fee = -1.0
	sumAmountTemp := 0.0
	countInUnspents := 0
	feePerVByte := getFeePerVByte()
	for i, unspent := range opReturn.Unspents {
		if unspent.Confirmations < opReturn.Confirmations {
			continue
		}
		countInUnspents += 1
		opReturn.Unspents[i].Expected = true

		sumAmountTemp += unspent.Amount

		// case 1.
		// Balance is 0, so did not need balance_tx
		tCountTxOuts := 1 + countExtra //  1(opreturn_data_tx) + extra_tx
		tFee := calFee(countInUnspents, tCountTxOuts, feePerVByte, opReturn.Address)
		if sumAmountTemp == tFee+payValueExtra {
			opReturn.Fee = tFee
			break
		}

		// case 2.
		tCountTxOuts = 1 + countExtra + 1 //  1(opreturn_data_tx) + extra_tx + 1(balance_tx)
		tFee = calFee(countInUnspents, tCountTxOuts, feePerVByte, opReturn.Address)
		if sumAmountTemp >= tFee+payValueExtra {
			opReturn.Fee = tFee
			break
		}
	}
	if opReturn.Fee <= 0.0 {
		err = fmt.Errorf("opReturn.Fee <= 0.0: not sufficient: sumUnspentAmount[%f] < fee[%f] + payValueExtra[%f]", sumAmountTemp, opReturn.Fee, payValueExtra)
		return
	}

	opReturn.AmountBalanceUsedUnspends = math.Round((sumAmountTemp-opReturn.Fee-payValueExtra)*100000000) / 100000000
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

	if opReturn.PayInfos == nil {
		opReturn.PayInfos = make(map[string]float64)
	}

	// 1. ListUnspent
	listUnspents, err := bitcoinCli.ListUnspentOfAddress(0, 0, []string{opReturn.Address})
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.ListUnspentOfAddress('%s'): %v", opReturn.Address, err)
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
		err = fmt.Errorf("@opReturn.selectUnspentsForSend(): %v", err)
		return
	}

	// 5. convertTextToHex
	opReturn.MessageHex = ConvertTextToHex(opReturn.Message)
	if err != nil {
		err = fmt.Errorf("@convertTextToHex('%s'): %v", opReturn.Message, err)
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

	if opReturn.AmountBalanceUsedUnspends > 0.00000000 {
		opReturn.PayInfos[opReturn.Address] = opReturn.AmountBalanceUsedUnspends // add balance-pay-info
	}
	opReturn.RawTx, err = bitcoinCli.CreateRawTransaction(createTxUnSpents, opReturn.PayInfos, opReturn.MessageHex)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.CreateRawTransaction(createTxUnSpents, opReturn.PayInfos, opReturn.MessageHex): %v", err)
		return
	}

	// 7. DumpPrivateKey
	if opReturn.PrivKey == "" {
		opReturn.PrivKey, err = bitcoinCli.DumpPrivateKey(opReturn.Address)
		if err != nil {
			err = fmt.Errorf("@bitcoinCli.DumpPrivateKey('%s'): %v", opReturn.Address, err)
			return
		}
	}

	// 8. SignRawTransactionWithKey
	opReturn.SignedRawTx, err = bitcoinCli.SignRawTransactionWithKey(opReturn.RawTx, opReturn.PrivKey)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.SignRawTransactionWithKey(opReturn.RawTx, opReturn.PrivKey): %v", err)
		return
	}

	// 9. SendRawTransaction
	opReturn.OpRetrunTxID, err = bitcoinCli.SendRawTransaction(opReturn.SignedRawTx)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.SendRawTransaction(opReturn.SignedRawTx): %v", err)
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
	listUnspents, err := bitcoinCli.ListUnspentOfAddress(0, 0, []string{payment.Address})
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.ListUnspentOfAddress('%s'): %v", payment.Address, err)
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
		err = fmt.Errorf("@payment.selectUnspentsForSend(): %v", err)
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
		err = fmt.Errorf("@bitcoinCli.CreateRawTransaction(createTxUnSpents, payment.PayInfos, ''): %v", err)
		return
	}

	// 7. DumpPrivateKey
	if payment.PrivKey == "" {
		payment.PrivKey, err = bitcoinCli.DumpPrivateKey(payment.Address)
		if err != nil {
			err = fmt.Errorf("@bitcoinCli.DumpPrivateKey('%s'): %v", payment.Address, err)
			return
		}
	}

	// 8. SignRawTransactionWithKey
	payment.SignedRawTx, err = bitcoinCli.SignRawTransactionWithKey(payment.RawTx, payment.PrivKey)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.SignRawTransactionWithKey(payment.RawTx, payment.PrivKey): %v", err)
		return
	}

	// 9. SendRawTransaction
	payment.PaymentTxID, err = bitcoinCli.SendRawTransaction(payment.SignedRawTx)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.SendRawTransaction(payment.SignedRawTx): %v", err)
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
		if tAmount < 0.0 { // means case[all of balance amount].
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
		err = fmt.Errorf("incorrect payment.PayInfos: %+v", payment.PayInfos)
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
	feePerVByte := getFeePerVByte()
	for i, unspent := range payment.Unspents {

		if unspent.Confirmations < payment.Confirmations {
			payment.Unspents[i].Expected = false
			continue
		}

		payment.Unspents[i].Expected = true
		sumSelectedUnspentsAmount += unspent.Amount
		countSelectedUnspents += 1
		payment.Fee = calFee(countSelectedUnspents, countPayment, feePerVByte, payment.Address)
		if sumSelectedUnspentsAmount >= payment.Fee+sumPaymentAmount {
			validSelectedUnspents = true
			if !hasTotalAmountCase { // for all of balance-amount
				break
			}
		}
	}

	if !validSelectedUnspents {
		err = fmt.Errorf("validSelectedUnspents is false: not sufficient: sumSelectedUnspentsAmount[%f] < fee[%f]+sumPaymentAmount[%f]", sumSelectedUnspentsAmount, payment.Fee, sumPaymentAmount)
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

type OpReturnReadable struct {
	BlockHash string
	BlockTime int64
	TxID      string
	Addresses []string
	Valid     bool
	Hex       string `json:",omitempty"`
	Readable  string `json:",omitempty"`
}

type OpReturnReadables struct {
	RpcUser    string
	RpcPW      string
	RpcConnect string
	RpcPort    string
	RpcPath    string
	Readables  []OpReturnReadable
}

func (opReturnReadables *OpReturnReadables) RunInBlockNum(blockNum int64) (err error) {

	bitcoinCli := goBitcoinCli.BitcoinRpc{
		RpcUser:    opReturnReadables.RpcUser,
		RpcPW:      opReturnReadables.RpcPW,
		RpcConnect: opReturnReadables.RpcConnect,
		RpcPort:    opReturnReadables.RpcPort,
		RpcPath:    opReturnReadables.RpcPath,
	}

	blockHash, err := bitcoinCli.GetBlockHash(blockNum)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.GetBlockHash(%d): %v", blockNum, err)
		return
	}

	err = opReturnReadables.RunInBlockHash(blockHash)
	if err != nil {
		err = fmt.Errorf("@opReturnReadables.RunInBlockHash(%s): %v", blockHash, err)
		return
	}

	return
}

func (opReturnReadables *OpReturnReadables) RunInBlockHash(blockHash string) (err error) {

	bitcoinCli := goBitcoinCli.BitcoinRpc{
		RpcUser:    opReturnReadables.RpcUser,
		RpcPW:      opReturnReadables.RpcPW,
		RpcConnect: opReturnReadables.RpcConnect,
		RpcPort:    opReturnReadables.RpcPort,
		RpcPath:    opReturnReadables.RpcPath,
	}

	block, err := bitcoinCli.GetBlock(blockHash)
	if err != nil {
		err = fmt.Errorf("@bitcoinCli.GetBlock(%s): %v", blockHash, err)
		return
	}
	txids := block["tx"].([]string)

	err = opReturnReadables.RunInTxIDs(txids)
	if err != nil {
		err = fmt.Errorf("@opReturnReadables.RunInTxIDs(txids): txids_%v: %v", txids, err)
		return
	}
	return
}

func (opReturnReadables *OpReturnReadables) RunInTxIDs(txids []string, onlyShowOpReturnTxIDs ...bool) (err error) {

	bitcoinCli := goBitcoinCli.BitcoinRpc{
		RpcUser:    opReturnReadables.RpcUser,
		RpcPW:      opReturnReadables.RpcPW,
		RpcConnect: opReturnReadables.RpcConnect,
		RpcPort:    opReturnReadables.RpcPort,
		RpcPath:    opReturnReadables.RpcPath,
	}

	onlyShowValid := true // default
	if len(onlyShowOpReturnTxIDs) > 0 {
		onlyShowValid = onlyShowOpReturnTxIDs[0]
	}

	opReturnReadables.Readables = make([]OpReturnReadable, 0)
	for _, txid := range txids {
		opReturnReadables.Readables = append(opReturnReadables.Readables, OpReturnReadable{TxID: txid})
	}

	readableRecord := make([]OpReturnReadable, 0)
	for _, record := range opReturnReadables.Readables {
		rawTxInfo, errI := bitcoinCli.GetRawTransaction(record.TxID)
		if errI != nil {
			continue
		}

		record.Addresses = make([]string, 0)
		for _, vinInfo := range rawTxInfo["vin"].([]map[string]interface{}) {

			vinRawTxInfo, errII := bitcoinCli.GetRawTransaction(vinInfo["txid"].(string))
			if errII != nil {
				continue
			}

			for _, vinVoutInfo := range vinRawTxInfo["vout"].([]map[string]interface{}) {
				vinVoutInfoN := vinVoutInfo["n"].(int)
				if vinInfo["vout"].(int) != vinVoutInfoN {
					continue
				}

				var scriptPubKeyInfo map[string]string
				inrec, errIII := json.Marshal(vinVoutInfo["scriptPubKey"])
				if errIII != nil {
					continue
				}
				json.Unmarshal(inrec, &scriptPubKeyInfo)
				record.Addresses = append(record.Addresses, scriptPubKeyInfo["address"])
			}
		}

		record.BlockHash = rawTxInfo["blockhash"].(string)
		record.BlockTime = rawTxInfo["blocktime"].(int64)
		record.Valid = false
		for _, voutInfo := range rawTxInfo["vout"].([]map[string]interface{}) {
			var scriptPubKeyInfo map[string]string
			inrec, errII := json.Marshal(voutInfo["scriptPubKey"])
			if errII != nil {
				continue
			}
			json.Unmarshal(inrec, &scriptPubKeyInfo)
			asmStr := scriptPubKeyInfo["asm"]
			if len(asmStr) >= 9 && asmStr[0:9] == "OP_RETURN" {
				record.Valid = true
				record.Hex = strings.Split(asmStr, "OP_RETURN ")[1]
				record.Readable, _ = ConvertHexToText(record.Hex)
			}
		}
		if !(onlyShowValid && !record.Valid) {
			readableRecord = append(readableRecord, record)
		}
	}
	opReturnReadables.Readables = readableRecord

	return
}

func getFeePerVByte() (fee float64) {

	fee = 4.0 // default
	minLimitFee := 3.0
	maxLimitFee := 25.0

	remoteFee, err := remoteFeePerVByte()

	if err != nil {
		return // default
	}

	if remoteFee <= 0.0 {
		return // default
	}
	if remoteFee <= minLimitFee {
		fee = minLimitFee
		return // min
	}
	if remoteFee >= maxLimitFee {
		fee = maxLimitFee
		return // max
	}

	fee = remoteFee
	return
}

func remoteFeePerVByte() (remoteFee float64, err error) {

	type CurrentFee struct {
		FastestFee  float64 `json:"fastestFee"`
		HalfHourFee float64 `json:"halfHourFee"`
		HourFee     float64 `json:"hourFee"`
		EconomyFee  float64 `json:"economyFee"`
		MinimumFee  float64 `json:"minimumFee"`
	}

	tURL := "https://mempool.space/api/v1/fees/recommended"
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1300) // TimeOut 1.3(sec)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tURL, nil)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	cFee := CurrentFee{}
	err = json.NewDecoder(res.Body).Decode(&cFee)
	if err != nil {
		return
	}
	// fmt.Println(cFee.FastestFee, cFee.HalfHourFee, cFee.HourFee, cFee.EconomyFee, cFee.MinimumFee)

	remoteFee = cFee.HalfHourFee
	return
}
