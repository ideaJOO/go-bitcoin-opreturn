package gobitcoinopreturn

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPayment(t *testing.T) {

	payment := Payment{}
	payment.RpcUser = "ideajoo"
	payment.RpcPW = "ideajoo123"
	payment.RpcConnect = "127.0.0.1"
	payment.RpcPath = fmt.Sprintf("wallet/%s", "test_07")
	payment.RpcPort = "18332"
	payment.Address = "tb1qvgucupwvjs5gjr4adljstwvak6749hqmztzye9" // 0.0010521

	payment.PayInfos = make(map[string]float64)

	payment.PayInfos["tb1qmhqe8pr06v0mefelardj4h6hkq095e5dh72mv3"] = 0.0001
	payment.PayInfos["tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"] = 0.0001
	payment.PayInfos["tb1qtc7nhjtqkkghvzc62gxf2crjf6fd9jde007juu"] = -1

	err := payment.Run()
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(payment)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestOpRetrun(t *testing.T) {

	opReturn := OpReturn{}
	opReturn.RpcUser = "ideajoo"
	opReturn.RpcPW = "ideajoo123"
	opReturn.RpcConnect = "127.0.0.1"
	opReturn.RpcPath = fmt.Sprintf("wallet/%s", "test_07")
	opReturn.RpcPort = "18332"
	opReturn.Address = "tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"
	opReturn.Message = "HELLO ideajoo/go-bitcoin-opreturn aaaa"

	err := opReturn.Run()
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturn)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestConvertHex(t *testing.T) {

	opReturn := OpReturn{}
	opReturn.RpcUser = "ideajoo"
	opReturn.RpcPW = "ideajoo123"
	opReturn.RpcConnect = "127.0.0.1"
	opReturn.RpcPort = "18332"
	opReturn.Address = "tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"
	opReturn.Message = "HELLO ideajoo/go-bitcoin-cli-light"

	err := opReturn.convertTextToHex()
	if err != nil {
		return
	}
	println()
	println(opReturn.MessageHex)
	println()
}
