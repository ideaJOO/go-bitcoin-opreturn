package gobitcoinopreturn

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCalFee(t *testing.T) {
	calFee(1, 2, 40, "bc1q")
}

func TestPayment(t *testing.T) {

	payment := Payment{}
	payment.RpcUser = "ideajoo"
	payment.RpcPW = "ideajoo123"
	payment.RpcConnect = "127.0.0.1"
	payment.RpcPath = fmt.Sprintf("wallet/%s", "satoshibento_order_id")
	payment.RpcPort = "8332"
	payment.Address = "17M94TyjrY832rDgyY4cn92qSf697LtWgS" // 0.0010521

	payment.PayInfos = make(map[string]float64)
	// 17M94TyjrY832rDgyY4cn92qSf697LtWgS
	payment.PayInfos["bc1qr3ypk033x9yeqwzfaczd98vckspf22v3nvw73c"] = -1
	// payment.PayInfos["tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"] = 0.0001
	// payment.PayInfos["tb1qtc7nhjtqkkghvzc62gxf2crjf6fd9jde007juu"] = -1

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
	opReturn.RpcPath = fmt.Sprintf("wallet/%s", "satoshibento_opreturn")
	opReturn.RpcPort = "8332"
	opReturn.Address = "bc1qr3ypk033x9yeqwzfaczd98vckspf22v3nvw73c"
	opReturn.Message = "ios/android App📱\nSatoshiPen:WriteOpReturn✏️\nSatoshiBook:ReadOpReturn📖"
	opReturn.LimitFeePerVByte = 50

	opReturn.PayInfos = make(map[string]float64)

	// opReturn.PayInfos["1EfzPvwXiTH9UeRDUeMCSBHFWhSejKQbWT"] = 0.00001000
	// payment.PayInfos["tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"] = 0.0001
	// payment.PayInfos["tb1qtc7nhjtqkkghvzc62gxf2crjf6fd9jde007juu"] = -1

	err := opReturn.Run()
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturn)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestConvertHexToUTF8(t *testing.T) {
	readable, isUTF8, _ := ConvertHexToText("f09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98adf09f98ad20202040717561696e7472656c6c65786274")
	fmt.Printf("\n===\n[%t]\n%s\n", isUTF8, readable)

	readable, isUTF8, _ = ConvertHexToText("58325bf65d66c74d5bf0cfe8092d7958a62243045abd562802aec51960f27b6c00fb6bf87d5b8c282c286f36d9216ed304e381995c6e239e245b7b7743c39368bec312000c05d10001000c048500503a")
	fmt.Printf("\n===\n[%t]\n%s\n", isUTF8, readable)

	readable, isUTF8, _ = ConvertHexToText("ec9588eb8595ed9598ec84b8ec9a940a68656c6c6f20f09f918b200a407361746f73686970656e")
	fmt.Printf("\n===\n[%t]\n%s\n", isUTF8, readable)
}

func TestConvertHex(t *testing.T) {

	message := "HELLO ideajoo/go-bitcoin-cli-light"
	message = "안녕하세요"
	message = "こんにちは 안녕하세요 สวัสดีค่ะ "

	println(message)
	hexStr := ConvertTextToHex(message)
	println(hexStr)

	// hexStr = "e38193e38293e381abe381a1e381af"
	readable, isUTF8, _ := ConvertHexToText(hexStr)
	fmt.Printf("\n[%t] %s\n", isUTF8, readable)
	return
	println("==============")
	println("==============")
	println("==============")
	println("==============")

	message =
		`@satoshibento  
-'SatBtI': Digest for Integrity
-'SatBtRecpt': CID for Receipt`
	hexStr = ConvertTextToHex(message)
	println(hexStr)
	println("407361746f73686962656e746f20200a2d27536174427449273a2044696765737420666f7220496e746567726974790a2d2753617442745265637074273a2043494420666f722052656365697074")

	readable, _, _ = ConvertHexToText(hexStr)
	println("S@" + strings.ReplaceAll(readable, " ", "!") + "@E")

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
		false,
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

	err := opReturnRecords.RunInBlockNum(2376040)
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

	err := opReturnRecords.RunInBlockHash("0000000081bfd8e5c0bf200f89bd2cf4be19816469624a4a09096edda5927ccc")
	if err != nil {
		fmt.Printf("\n\n\n%+v\n\n", err)
		return
	}

	jsonDump, _ := json.Marshal(opReturnRecords)
	fmt.Printf("\n\n\n%+v\n\n", string(jsonDump))
}

func TestCurrentlyFee(t *testing.T) {

	rFee := RemoteFees{}
	rFee.remoteFeePerVByte2()
	fmt.Printf("\n\n%v", rFee)
	index := 0
	for {
		if index > 5 {
			break
		}
		index += 1
		fmt.Printf("\n%v\n", getFeePerVByte2(50))
	}

}
