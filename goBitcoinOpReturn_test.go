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
	opReturn.ReceiptText = "HELLO ideajoo/go-bitcoin-opreturn "

	err := opReturn.Run()
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
