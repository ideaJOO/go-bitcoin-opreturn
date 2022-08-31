package gobitcoinopreturn

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
)

type OpReturn struct {
	RpcUser                   string
	RpcPW                     string
	RpcConnect                string
	RpcPort                   string
	Address                   string
	PrivKey                   string
	ReceiptText               string
	ReceiptHex                string
	Unspents                  []Unspent
	AmountUnspends            float64
	Fee                       float64
	AmountBalanceUsedUnspends float64
	RawTx                     string
	SignedRawTx               string
	OpRetrunTxID              string
}

type Unspent struct {
	TxID     string
	Vout     int
	Amount   float64
	Expected bool
}

type JsonRpc struct {
	JsonRpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

func (opReturn OpReturn) requestJsonRpc(jsonRpcBytes []byte) (body []byte, err error) {

	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/", opReturn.RpcConnect, opReturn.RpcPort), bytes.NewBuffer(jsonRpcBytes))
	if err != nil {
		err = fmt.Errorf("@requestJsonRpc: %s", err)
		return
	}
	request.Header.Set("content-type", "text/plain;")
	request.SetBasicAuth(opReturn.RpcUser, opReturn.RpcPW)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("@requestJsonRpc: %s", err)
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("@requestJsonRpc: %s", err)
		return
	}
	return
}

// ListUnspentOfAddress
func (opReturn *OpReturn) ListUnspentOfAddress() (err error) {

	type listUnspentInfo struct {
		TxID          string   `json:"txid"`          // (string) the transaction id
		Vout          int      `json:"vout"`          // (numeric) the vout value
		Address       string   `json:"address"`       // (string) the bitcoin address
		Label         string   `json:"label"`         // (string) The associated label, or "" for the default label
		ScriptPutKey  string   `json:"scriptPubKey"`  // (string) the script key
		Amount        float64  `json:"amount"`        // (numeric) the transaction output amount in BTC
		Confirmations int      `json:"confirmations"` // (numeric) The number of confirmations
		RedeemScript  string   `json:"redeemScript"`  // (string) The redeemScript if scriptPubKey is P2SH
		WitnessScript string   `json:"witnessScript"` // (string) witnessScript if the scriptPubKey is P2WSH or P2SH-P2WSH
		Spendable     bool     `json:"spendable"`     // (boolean) Whether we have the private keys to spend this output
		Solvable      bool     `json:"solvable"`      // (boolean) Whether we know how to spend this output, ignoring the lack of keys
		Reused        bool     `json:"reused"`        // (boolean) (only present if avoid_reuse is set) Whether this output is reused/dirty (sent to an address that was previously spent from)
		Desc          string   `json:"desc"`          // (string) (only when solvable) A descriptor for spending this output
		ParentDescs   []string `json:"parent_descs"`  //
		Safe          bool     `json:"safe"`          // (boolean) Whether this output is considered safe to spend. Unconfirmed transactions
	}

	jsonRpcBytes, err := json.Marshal(JsonRpc{
		JsonRpc: "1.0",
		ID:      "GoBitcoinOpReturn",
		Method:  "listunspent",
		Params:  []interface{}{1, 9999999, []string{opReturn.Address}},
	})
	if err != nil {
		err = fmt.Errorf("@ListUnspentOfAddress: %s", err)
		return
	}

	body, err := opReturn.requestJsonRpc(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@ListUnspentOfAddress: %s", err)
		return
	}

	type resultListUnspent struct {
		ListUnspent []listUnspentInfo `json:"result"`
	}
	result := resultListUnspent{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@ListUnspentOfAddress(): %s", err)
		return
	}
	opReturn.Unspents = make([]Unspent, 0)
	for _, listUnspentInfo := range result.ListUnspent {
		opReturn.Unspents = append(opReturn.Unspents, Unspent{
			TxID:     listUnspentInfo.TxID,
			Vout:     listUnspentInfo.Vout,
			Amount:   math.Round(listUnspentInfo.Amount*100000000) / 100000000,
			Expected: false,
		})
	}
	sort.Slice(opReturn.Unspents, func(i, j int) bool {
		return opReturn.Unspents[i].Amount > opReturn.Unspents[j].Amount
	})

	return
}

func (opReturn *OpReturn) calAmountUnspents() (err error) {
	balance := 0.0
	for _, unspent := range opReturn.Unspents {
		balance += unspent.Amount
	}
	opReturn.AmountUnspends = math.Round(balance*100000000) / 100000000
	return
}

func (opReturn *OpReturn) convertTextToHex() (err error) {
	opReturn.ReceiptHex = hex.EncodeToString([]byte(opReturn.ReceiptText))
	return
}

func (opReturn *OpReturn) CalFee() (err error) {

	// TODO: Cal fee by bytes
	fee := 0.00010000 // Temp Fee

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
		if unspent.Amount > 0.0 && sumAmountTemp >= opReturn.Fee {
			continue
		}
		opReturn.Unspents[i].Expected = true
		sumAmountTemp += unspent.Amount
		sumAmountTemp = math.Round((sumAmountTemp)*100000000) / 100000000
	}
	opReturn.AmountBalanceUsedUnspends = math.Round((sumAmountTemp-opReturn.Fee)*100000000) / 100000000
	return
}

func (opReturn *OpReturn) createRawTransaction() (err error) {

	if err = opReturn.CalFee(); err != nil {
		err = fmt.Errorf("@CreateRawTransaction(): %s", err)
		return
	}

	if err = opReturn.selectUnspentsForSend(); err != nil {
		err = fmt.Errorf("@CreateRawTransaction(): %s", err)
		return
	}

	if err = opReturn.convertTextToHex(); err != nil {
		err = fmt.Errorf("@CreateRawTransaction(): %s", err)
		return
	}

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

	tCreateTxData := make(map[string]interface{})
	tCreateTxData["data"] = opReturn.ReceiptHex
	tCreateTxData[opReturn.Address] = opReturn.AmountBalanceUsedUnspends

	jsonRpcBytes, err := json.Marshal(JsonRpc{
		JsonRpc: "1.0",
		ID:      "GoBitcoinOpReturn",
		Method:  "createrawtransaction",
		Params:  []interface{}{createTxUnSpents, tCreateTxData},
	})
	if err != nil {
		err = fmt.Errorf("@CreateRawTransaction: %s", err)
		return
	}

	body, err := opReturn.requestJsonRpc(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@CreateRawTransaction: %s", err)
		return
	}

	type resultCreateRaxTx struct {
		RawTx string `json:"result"`
	}
	result := resultCreateRaxTx{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@CreateRawTransaction(): %s", err)
		return
	}

	opReturn.RawTx = result.RawTx

	return
}

func (opReturn *OpReturn) dumpPrivateKey() (err error) {

	jsonRpcBytes, err := json.Marshal(JsonRpc{
		JsonRpc: "1.0",
		ID:      "GoBitcoinOpReturn",
		Method:  "dumpprivkey",
		Params:  []interface{}{opReturn.Address},
	})
	if err != nil {
		err = fmt.Errorf("@dumpPrivateKey: %s", err)
		return
	}

	body, err := opReturn.requestJsonRpc(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@dumpPrivateKey: %s", err)
		return
	}

	type resultDumpPrivateKey struct {
		PrivKey string `json:"result"`
	}
	result := resultDumpPrivateKey{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@dumpPrivateKey(): %s", err)
		return
	}

	opReturn.PrivKey = result.PrivKey

	return
}

func (opReturn *OpReturn) SignRawTransactionWithKey() (err error) {
	if opReturn.PrivKey == "" {
		err = opReturn.dumpPrivateKey()
		if err != nil {
			err = fmt.Errorf("@SignRawTransactionWithKey(): %s", err)
			return
		}
	}

	jsonRpcBytes, err := json.Marshal(JsonRpc{
		JsonRpc: "1.0",
		ID:      "GoBitcoinOpReturn",
		Method:  "signrawtransactionwithkey",
		Params:  []interface{}{opReturn.RawTx, []string{opReturn.PrivKey}},
	})
	if err != nil {
		err = fmt.Errorf("@SignRawTransactionWithKey(): %s", err)
		return
	}

	body, err := opReturn.requestJsonRpc(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@SignRawTransactionWithKey(): %s", err)
		return
	}

	type signedRawTxInfo struct {
		Hex      string `json:"hex"`      // (string) the transaction id
		Complete bool   `json:"complete"` // (numeric) the vout value
	}
	type resultSignedRawTxInfo struct {
		SignedRawTxInfo signedRawTxInfo `json:"result"`
	}

	tmpBody := string(body)
	fmt.Printf("%s", tmpBody)

	result := resultSignedRawTxInfo{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@SignRawTransactionWithKey(): %s", err)
		return
	}

	opReturn.SignedRawTx = result.SignedRawTxInfo.Hex

	return
}

func (opReturn *OpReturn) SendRawTransaction() (err error) {

	jsonRpcBytes, err := json.Marshal(JsonRpc{
		JsonRpc: "1.0",
		ID:      "GoBitcoinOpReturn",
		Method:  "sendrawtransaction",
		Params:  []interface{}{opReturn.SignedRawTx},
	})
	if err != nil {
		err = fmt.Errorf("@SendRawTransaction: %s", err)
		return
	}

	body, err := opReturn.requestJsonRpc(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@SendRawTransaction: %s", err)
		return
	}

	type resultSendRawTx struct {
		OpReturnTxID string `json:"result"`
	}
	result := resultSendRawTx{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@SendRawTransaction(): %s", err)
		return
	}
	opReturn.OpRetrunTxID = result.OpReturnTxID
	return
}
