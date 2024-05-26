package goqris

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	mathrand "math/rand"
	"os"
	"strconv"
	"time"
)

type BriMpmDynamic struct {
	Host         string
	ClientID     string
	ClientSecret string
	PartnerID    string
	PrivateKey   string
	MerchantID   string
	TerminalID   string
	ChannelID    string
	Timezone     string
}

func (mpm BriMpmDynamic) getPrivateKey() (*rsa.PrivateKey, error) {
	c, err := os.ReadFile(mpm.PrivateKey)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(c)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func (mpm BriMpmDynamic) sha256(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	return h.Sum(nil)
}

func (mpm BriMpmDynamic) hmac(s string) string {
	mac := hmac.New(sha512.New, []byte(mpm.ClientSecret))
	mac.Write([]byte(s))
	return hex.EncodeToString(mac.Sum(nil))
}

func (mpm BriMpmDynamic) getSignature(s string) (string, error) {
	bs := mpm.sha256([]byte(s))
	pk, err := mpm.getPrivateKey()
	if err != nil {
		return "", err
	}

	sign, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, bs)
	return base64.StdEncoding.EncodeToString(sign), err
}

func (mpm BriMpmDynamic) getExternalId() string {
	max := 99999999999
	min := 9999999999
	r := mathrand.Intn(max-min) + min
	return strconv.FormatInt(int64(r), 10)
}

func (mpm BriMpmDynamic) getTimestamp() (string, error) {
	now := time.Now()
	if mpm.Timezone != "" {
		tz, err := time.LoadLocation(mpm.Timezone)
		if err != nil {
			return "", err
		}

		now = now.In(tz)
	}

	return now.Format(time.RFC3339), nil
}

// example request:
//
//	{
//		"grantType": "client_credentials"
//	}
//
// example response:
//
//	{
//	  "accessToken": "jwy7GgloLqfqbZ9OnxGxmYOuGu85",
//	  "tokenType": "BearerToken",
//	  "expiresIn": "899"
//	}
func (mpm BriMpmDynamic) getAccessToken() (string, error) {
	cacheKey := "briMpm"
	if accessToken := getAccessTokenFromCache(cacheKey); accessToken != "" {
		return accessToken, nil
	}

	uri := fmt.Sprintf("%s/%s", mpm.Host, "snap/v1.0/access-token/b2b")
	bodyMap := M{"grantType": "client_credentials"}
	timestamp, err := mpm.getTimestamp()
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf("%s|%s", mpm.ClientID, timestamp)
	signature, err := mpm.getSignature(message)
	if err != nil {
		return "", err
	}

	headers := M{
		"Content-Type": "application/json",
		"X-CLIENT-KEY": mpm.ClientID,
		"X-TIMESTAMP":  timestamp,
		"X-SIGNATURE":  signature,
	}
	data := makeRequest(uri, bodyMap, headers)
	if data.Err != nil {
		return "", data.Err
	}

	if at := data.Data.GetValue("accessToken"); at == "" {
		errorCode := data.Data.GetValue("responseCode")
		errorDesc := data.Data.GetValue("responseMessage")
		return "", fmt.Errorf("BRIMpmDynamic(%s) - %s", errorCode, errorDesc)
	}

	return setAccessTokenToCache(cacheKey, data.Data)
}

// example request:
//
//	{
//		"partnerReferenceNo": "10009121031000912103",
//		"amount": {
//			"value": "15001.00",
//			"currency": "IDR"
//	     },
//		"merchantId": "1234567890",
//		"terminalId": "10049258"
//	}
//
// example response:
//
//	{
//	  "responseCode": "2004700",
//	  "responseMessage": "Successful",
//	  "referenceNo": "000008526955",
//	  "partnerReferenceNo": "10009121031000912103",
//	  "qrContent": "00020101021226650013ID.CO.BRI.WWW011893600002021046147202150000010190000140303UME520451115303360540450005802ID5919Ritual Kopi Bandung6005BERAU6105773126222011812606343410585824163049C2F"
//	}
func (mpm BriMpmDynamic) GenerateQRCode(r IRequest) (M, error) {
	accessToken, err := mpm.getAccessToken()
	if err != nil {
		return nil, err
	}

	bodyMap := r.Payload()
	bodyMap.SetValueIfEmpty("merchantId", mpm.MerchantID)
	bodyMap.SetValueIfEmpty("terminalId", mpm.TerminalID)
	bodyBytes, _ := json.Marshal(bodyMap)
	bodyHex := hex.EncodeToString(mpm.sha256(bodyBytes))
	endpoint := "v1.0/qr-dynamic-mpm/qr-mpm-generate-qr"
	timestamp, err := mpm.getTimestamp()
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("POST:/%s:%s:%s:%s", endpoint, accessToken, bodyHex, timestamp)
	headers := M{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accessToken,
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   mpm.hmac(message),
		"X-PARTNER-ID":  mpm.PartnerID,
		"X-EXTERNAL-ID": mpm.getExternalId(),
		"CHANNEL-ID":    mpm.ChannelID,
	}

	uri := fmt.Sprintf("%s/%s", mpm.Host, endpoint)
	data := makeRequest(uri, bodyMap, headers)
	return data.Data, data.Err
}

// example request:
//
//	{
//		"originalReferenceNo": "000008526955",
//		"serviceCode": "17",
//		"additionalInfo": {
//			"terminalId": "10049258"
//	    }
//	}
//
// example response:
//
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
func (mpm BriMpmDynamic) CheckStatusTransaction(r IRequest) (M, error) {
	accessToken, err := mpm.getAccessToken()
	if err != nil {
		return nil, err
	}

	bodyMap := r.Payload()
	bodyMap.SetValueIfEmptyWithFunc("additionalInfo", func(m M) M {
		m.SetValueIfEmpty("terminalId", mpm.TerminalID)
		return m
	})
	bodyBytes, _ := json.Marshal(bodyMap)
	hexBody := hex.EncodeToString(mpm.sha256(bodyBytes))
	timestamp, err := mpm.getTimestamp()
	if err != nil {
		return nil, err
	}

	endpoint := "v1.0/qr-dynamic-mpm/qr-mpm-query"
	message := fmt.Sprintf("POST:/%s:%s:%s:%s", endpoint, accessToken, hexBody, timestamp)
	headers := M{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accessToken,
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   mpm.hmac(message),
		"X-PARTNER-ID":  mpm.PartnerID,
		"X-EXTERNAL-ID": mpm.getExternalId(),
		"CHANNEL-ID":    mpm.ChannelID,
	}

	uri := fmt.Sprintf("%s/%s", mpm.Host, endpoint)
	data := makeRequest(uri, bodyMap, headers)
	return data.Data, data.Err
}

type BriMpmDynamicGenerateQRCodeRequest struct {
	PartnerReferenceNo string
	Amount             string
	Currency           string
}

func (r BriMpmDynamicGenerateQRCodeRequest) Payload() M {
	return M{
		"partnerReferenceNo": r.PartnerReferenceNo,
		"amount": M{
			"value":    r.Amount,
			"currency": r.Currency,
		},
	}
}

type BriMpmDynamicCheckStatusTransactionRequest struct {
	OriginalReferenceNo string
	ServiceCode         string
}

func (r BriMpmDynamicCheckStatusTransactionRequest) Payload() M {
	return M{
		"originalReferenceNo": r.OriginalReferenceNo,
		"serviceCode":         r.ServiceCode,
	}
}
