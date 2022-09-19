package gobitcoinopreturn

import (
	"fmt"
	"testing"
)

func TestListUnspentOfAddress(t *testing.T) {

	opReturn := OpReturn{}
	opReturn.RpcUser = "ideajoo"
	opReturn.RpcPW = "ideajoo123"
	opReturn.RpcConnect = "127.0.0.1"
	opReturn.RpcPort = "18332"
	opReturn.Address = "tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"
	opReturn.Message = "HELLO ideajoo/go-bitcoin-opreturn!!!!!"

	err := opReturn.Run("test_07")
	if err != nil {
		return
	}
	fmt.Printf("\n\n\n%+v\n\n", opReturn)

	// opReturn.calAmountUnspents()
	// fmt.Printf("\n\n\n%+v\n\n", opReturn)

	// opReturn.createRawTransaction()
	// fmt.Printf("\n\n\n%+v\n\n", opReturn)

	// opReturn.dumpPrivateKey()
	// fmt.Printf("\n\n\n%+v\n\n", opReturn)

	// opReturn.SignRawTransactionWithKey()
	// fmt.Printf("\n\n\n%+v\n\n", opReturn)

	// opReturn.SendRawTransaction()
	// fmt.Printf("\n\n\n%+v\n\n", opReturn)

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
