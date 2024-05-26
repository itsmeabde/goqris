package goqris

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

var briMpmDynamicTest = &BriMpmDynamic{
	Host:         "https://sandbox.partner.api.bri.co.id",
	ClientID:     "",
	ClientSecret: "",
	PrivateKey:   "",
	MerchantID:   "",
	TerminalID:   "",
	PartnerID:    "",
	ChannelID:    "",
	Timezone:     "",
}

func TestBriMmpDynamic(t *testing.T) {
	// create request for generate QRCode
	timestampStr := strconv.FormatInt(time.Now().Unix(), 10)
	suffix := "123"
	partnerReferenceNo := timestampStr + suffix

	// generate QRcode
	res, err := briMpmDynamicTest.GenerateQRCode(BriMpmDynamicGenerateQRCodeRequest{
		PartnerReferenceNo: partnerReferenceNo,
		Amount:             "5000.00",
		Currency:           "IDR",
	})
	if err != nil {
		t.Error(err)
		return
	}

	dataBytes, _ := json.Marshal(res)

	// check if request successful
	if successful := res.SuccessfulGenerate(); !successful {
		t.Error(string(dataBytes))
		return
	}

	t.Log(string(dataBytes))

	// get reference no -> required for check status transaction
	referenceNo := res.RefNo()
	// get service code -> required for check status transaction
	serviceCode := res.ServiceCode()

	// check status transaction
	res, err = briMpmDynamicTest.CheckStatusTransaction(BriMpmDynamicCheckStatusTransactionRequest{
		OriginalReferenceNo: referenceNo,
		ServiceCode:         serviceCode,
	})
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Paid: %v\n", res.SuccessfulPaid())

	dataBytes, _ = json.Marshal(res)
	t.Log(string(dataBytes))
}

var bniTest = &Bni{
	Host:         "https://mom-trxauth.spesandbox.com",
	Username:     "",
	Password:     "",
	ClientID:     "",
	ClientSecret: "",
	HmacKey:      "",
	MerchantID:   "",
	TerminalID:   "",
}

func TestBni(t *testing.T) {
	// create random orderID
	timestampStr := strconv.FormatInt(time.Now().Unix(), 10)
	requestID := timestampStr + "123"

	// create expired transaction
	layoutLocal := "2006-01-02T15:04:05"
	exp := time.Now().Add(1 * time.Hour).Format(layoutLocal)

	// generate QRcode
	res, err := bniTest.GenerateQRCode(BniGenerateQRCodeRequest{
		RequestID: requestID,
		Amount:    "5000.00",
		QRExpired: exp,
	})
	if err != nil {
		t.Error(err)
		return
	}

	dataBytes, _ := json.Marshal(res)

	// check if request successful
	if successful := res.SuccessfulGenerate(); !successful {
		t.Error(string(dataBytes))
		return
	}

	// get bill number -> required for check status transaction
	billNumber := res.RefNo()

	// check status transaction
	res, err = bniTest.CheckStatusTransaction(BniCheckStatusTransactionRequest{
		RequestID:  timestampStr,
		BillNumber: billNumber,
	})
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Paid: %v\n", res.SuccessfulPaid())

	dataBytes, _ = json.Marshal(res)
	t.Log(string(dataBytes))
}
