## About QRIS (Quick Response Code Indonesian Standard)

QRIS is a National QR code standard to facilitate QR code payments in Indonesia launched by Bank Indonesia and the Indonesian Payment System Association (ASPI).

## Installation
```shell
go get github.com/itsmeabde/goqris
```

## Quickstart
### BRI QRIS Merchant Presented Mode (MPM) Dinamis
Generate QRCode
```go
package main

import (
    "encoding/json"
    "log"

    "github.com/itsmeabe/goqris"
)

func main() {
    briMmpDynamic := &goqris.BriMpmDynamic{
	Host:         "https://sandbox.partner.api.bri.co.id",
	ClientID:     "*****",
	ClientSecret: "*****",
	PrivateKey:   "path/to/private.pem",
	MerchantID:   "*****",
	TerminalID:   "*****",
	PartnerID:    "*****",
	ChannelID:    "YOUR-CHANNEL-NAME",
	Timezone:     "CUSTOM-TIMEZONE",
    }

    req := goqris.BriMpmDynamicGenerateQRCodeRequest{
	PartnerReferenceNo: "YOUR-ID",
	Amount:             "15001.00",
	Currency:           "IDR",
    }

    res, err := briMmpDynamic.GenerateQRCode(req)
    if err != nil {
	panic(err)
    }

    jsonBytes, _ := json.Marshal(res)
    jsonStr := string(jsonBytes)

    // check if request is successful
    if successful := res.SuccessfulGenerate(); !successful {
	panic(jsonStr)
    }

    log.Println(jsonStr)

    //	{
    //	  "responseCode": "2004700",
    //	  "responseMessage": "Successful",
    //	  "referenceNo": "000008526955",
    //	  "partnerReferenceNo": "10009121031000912103",
    //	  "qrContent": "00020101021226650013ID.CO.BRI.WWW011893600002021046147202150000010190000140303UME520451115303360540450005802ID5919Ritual Kopi Bandung6005BERAU6105773126222011812606343410585824163049C2F"
    //	}

    // get reference no -> required for check status transaction
    referenceNo := res.RefNo()
    // get service code -> required for check status transaction
    serviceCode := res.ServiceCode()
}
```

Check status transaction
```go
package main

import (
    "encoding/json"
    "log"

    "github.com/itsmeabe/goqris"
)

func main() {
    briMmpDynamic := &goqris.BriMpmDynamic{
	Host:         "https://sandbox.partner.api.bri.co.id",
	ClientID:     "*****",
	ClientSecret: "*****",
	PrivateKey:   "path/to/private.pem",
	MerchantID:   "*****",
	TerminalID:   "*****",
	PartnerID:    "*****",
	ChannelID:    "YOUR-CHANNEL-NAME",
	Timezone:     "CUSTOM-TIMEZONE",
    }

    req := goqris.BriMpmDynamicCheckStatusTransactionRequest{
	OriginalReferenceNo: "000008526955",
	ServiceCode:         "47",
    }

    res, err := briMmpDynamic.CheckStatusTransaction(req)
    if err != nil {
	panic(err)
    }

    jsonBytes, _ := json.Marshal(res)
    jsonStr := string(jsonBytes)

    // check if request is successful
    if successful := res.SuccessfulPaid(); !successful {
	panic(jsonStr)
    }

    log.Println(jsonStr)

    //	{
    //	  "responseCode": "2005100",
    //	  "responseMessage": "Successful",
    //	  "originalReferenceNo": "000008526955",
    //	  "serviceCode": "17",
    //	  "latestTransactionStatus": "00",
    //	  "transactionStatusDesc": "Successfully",
    //	  "amount": {
    //	     "value": "15001.00",
    //	     "currency": "IDR"
    //	  },
    //	  "terminalId": "10049258",
    //	  "additionalInfo": {
    //	     "customerName": "John Doe",
    //	     "customerNumber": "9360015723456789",
    //	     "invoiceNumber": "10009121031000912103",
    //	     "issuerName": "Finnet 2",
    //	     "mpan": "9360000201102921379"
    //	  }
    //	}
}
```

### BNI
Generate QRCode
```go
package main

import (
    "encoding/json"
    "log"
    "time"

    "github.com/itsmeabe/goqris"
)

func main() {
    bni := &goqris.Bni{
	Host:         "https://mom-trxauth.spesandbox.com",
	Username:     "*****",
	Password:     "*****",
	ClientID:     "*****",
	ClientSecret: "*****",
	HmacKey:      "*****",
	MerchantID:   "*****",
	TerminalID:   "*****",
    }

    // create expired transaction
    layoutLocal := "2006-01-02T15:04:05"
    exp := time.Now().Add(1 * time.Hour).Format(layoutLocal)

    req := goqris.BniGenerateQRCodeRequest{
	RequestID: "YOUR-ID",
	Amount:    "10000.00",
	QRExpired: exp,
    }

    res, err := bni.GenerateQRCode(req)
    if err != nil {
	panic(err)
    }

    jsonBytes, _ := json.Marshal(res)
    jsonStr := string(jsonBytes)

    // check if request is successful
    if successful := res.SuccessfulGenerate(); !successful {
	panic(dataStr)
    }

    log.Println(jsonStr)

    //	{
    //	  "code": "00",
    //	  "message": "success",
    //	  "bill_number": "C000011957",
    //	  "nmid": "ID220614113906351",
    //	  "qr_string": "00020101021226590013ID.CO.BNI.WWW011893600009150002344302096579089700303UBE51470015ID.OR.GPNQR.WWW0217ID2206141139063510303UBE5204762353033605406100.005802ID5925PT Logika Garis Elektroni6007GIANYAR61058057162180110C00001195707006304709E",
    //	  "qr_expired": "2022-06-23T15:01:28"
    //	}

    // get bill number -> required for check status transaction
    billNumber := res.RefNo()
}
```

Check status transaction
```go
package main

import (
    "encoding/json"
    "log"
    "strconv"
    "time"

    "github.com/itsmeabe/goqris"
)

func main() {
    bni := &goqris.Bni{
	Host:         "https://mom-trxauth.spesandbox.com",
	Username:     "*****",
	Password:     "*****",
	ClientID:     "*****",
	ClientSecret: "*****",
	HmacKey:      "*****",
	MerchantID:   "*****",
	TerminalID:   "*****",
    }

    requestId := strconv.FormatInt(time.Now().Unix(), 10)
    req := goqris.BniCheckStatusTransactionRequest{
	RequestID:  requestId,
	BillNumber: "C000011957",
    }

    res, err := bni.CheckStatusTransaction(req)
    if err != nil {
	panic(err)
    }

    jsonBytes, _ := json.Marshal(res)
    jsonStr := string(jsonBytes)

    // check if request is successful
    if successful := res.SuccessfulPaid(); !successful {
	panic(jsonStr)
    }

    log.Println(jsonStr)

    //	{
    //	    "code": "00",
    //	    "message": "success",
    //	    "request_id": "XwVjF5zfuHhrDZuw",
    //	    "customer_pan": "9360001110000000019",
    //	    "amount": "10000.00",
    //	    "transaction_datetime": "2021-02-25T13:36:13",
    //	    "amount_fee": "1000.00",
    //	    "rrn": "123456789012",
    //	    "bill_number": "12345678901234567890",
    //	    "issuer_code": "93600013",
    //	    "customer_name": "John Doe",
    //	    "terminal_id": "00005771",
    //	    "merchant_id": "008800223497",
    //	    "stan": "210226",
    //	    "merchant_name": "Sukses Makmur Bendungan Hilir",
    //	    "approval_code": "00",
    //	    "merchant_pan": "936000131600000003",
    //	    "mcc": "5814",
    //	    "merchant_city": "Jakarta Pusat",
    //	    "merchant_country": "ID",
    //	    "currency_code": "360",
    //	    "payment_status": "00",
    //	    "payment_description": "Payment Success"
    //	}
}
```
