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
	opReturn.Message = "HELLO ideajoo/go-bitcoin-opreturn aaaa!!@"

	err := opReturn.Run()
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturn)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestConvertHex(t *testing.T) {

	message := "HELLO ideajoo/go-bitcoin-cli-light"

	hexStr := convertTextToHex(message)
	readable, _ := convertHexToText(hexStr)

	println(message)
	println(hexStr)
	println(readable)
	println()

}

func TestOpRetrunReadableTxIDs(t *testing.T) {

	opReturnRecords := OpReturnReadables{}
	opReturnRecords.RpcUser = "ideajoo"
	opReturnRecords.RpcPW = "ideajoo123"
	opReturnRecords.RpcConnect = "127.0.0.1"
	opReturnRecords.RpcPath = fmt.Sprintf("wallet/%s", "test_07")
	opReturnRecords.RpcPort = "18332"

	err := opReturnRecords.RunInTxIDs(
		[]string{"990612a08b69ea5c2b27c1754a728aa5745e87f4afd7589f41c6591d976cd44c", "48bcb4b44f4fca4b3d35ac4c10a16f9e4353b619ca6dc28a903ca09e4136fdfe"},
		// false,
	)
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturnRecords)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestOpRetrunRecordsBlockNumber(t *testing.T) {

	opReturnRecords := OpReturnReadables{}
	opReturnRecords.RpcUser = "ideajoo"
	opReturnRecords.RpcPW = "ideajoo123"
	opReturnRecords.RpcConnect = "127.0.0.1"
	opReturnRecords.RpcPath = fmt.Sprintf("wallet/%s", "test_07")
	opReturnRecords.RpcPort = "18332"

	err := opReturnRecords.RunInBlockNum(2377224)
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturnRecords)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestOpRetrunRecordsBlockHash(t *testing.T) {

	opReturnRecords := OpReturnReadables{}
	opReturnRecords.RpcUser = "ideajoo"
	opReturnRecords.RpcPW = "ideajoo123"
	opReturnRecords.RpcConnect = "127.0.0.1"
	opReturnRecords.RpcPath = fmt.Sprintf("wallet/%s", "test_07")
	opReturnRecords.RpcPort = "18332"

	err := opReturnRecords.RunInBlockHash("00000000a363fb45c720d87e0390787ecec2f3bcd62fdd3f97239fec9312d439")
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturnRecords)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}
