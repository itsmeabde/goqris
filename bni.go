package goqris

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type Bni struct {
	Host         string
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	HmacKey      string
	MerchantID   string
	TerminalID   string
}

func (bni Bni) getBasicToken() string {
	secret := fmt.Sprintf("%s:%s", bni.ClientID, bni.ClientSecret)
	basicToken := base64.StdEncoding.EncodeToString([]byte(secret))

	return fmt.Sprintf("Basic %s", basicToken)
}

func (bni Bni) getSignature(s string) string {
	mac := hmac.New(sha512.New, []byte(bni.HmacKey))
	mac.Write([]byte(s))
	sign := mac.Sum(nil)

	return hex.EncodeToString(sign)
}

// example request:
//
//	{
//		"grant_type": "password",
//		"username": "user",
//		"password": "secret"
//	}
//
// example response:
//
//	{
//	  "access_token": "jwy7GgloLqfqbZ9OnxGxmYOuGu85",
//	  "token_type": "Bearer",
//	  "expires_in": "899"
//	}
func (bni Bni) getAccessToken() (string, error) {
	cacheKey := "bni"
	if accessToken := getAccessTokenFromCache(cacheKey); accessToken != "" {
		return accessToken, nil
	}

	uri := fmt.Sprintf("%s/%s", bni.Host, "auth/get-token")
	bodyMap := M{
		"username":   bni.Username,
		"password":   bni.Password,
		"grant_type": "password",
	}
	headers := M{
		"Content-Type":  "application/json",
		"Authorization": bni.getBasicToken(),
	}
	data := makeRequest(uri, bodyMap, headers)
	if data.Err != nil {
		return "", data.Err
	}

	if _, ok := data.Data["access_token"]; !ok {
		errCode, _ := data.Data["code"].(string)
		errDesc, _ := data.Data["error"].(string)
		return "", fmt.Errorf("BNI(%s) - %s", errCode, errDesc)
	}

	return setAccessTokenToCache(cacheKey, data.Data)
}

// example request:
//
//	{
//		"request_id": "10009121031000912103",
//		"amount": "15001.00"
//		"merchant_id": "1234567890",
//		"terminal_id": "10049258",
//		"qr_expired": "2022-06-23T15:01:28"
//	}
//
// example response:
//
//	{
//	  "code": "00",
//	  "message": "success",
//	  "bill_number": "C000011957",
//	  "nmid": "ID220614113906351",
//	  "qr_string": "00020101021226590013ID.CO.BNI.WWW011893600009150002344302096579089700303UBE51470015ID.OR.GPNQR.WWW0217ID2206141139063510303UBE5204762353033605406100.005802ID5925PT Logika Garis Elektroni6007GIANYAR61058057162180110C00001195707006304709E",
//	  "qr_expired": "2022-06-23T15:01:28"
//	}
func (bni Bni) GenerateQRCode(r IRequest) (M, error) {
	accessToken, err := bni.getAccessToken()
	if err != nil {
		return nil, err
	}

	bodyMap := r.Payload()
	id := bodyMap.GetValue("request_id")
	exp := bodyMap.GetValue("qr_expired")
	message := fmt.Sprintf("%s:%s:%s", id, bni.MerchantID, exp)
	headers := M{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accessToken,
		"X-Signature":   bni.getSignature(message),
	}
	uri := fmt.Sprintf("%s/%s", bni.Host, "qr/generate-qr")
	data := makeRequest(uri, bodyMap, headers)
	return data.Data, data.Err
}

// example request
//
//	{
//		"request_id": "XwVjF5zfuHhrDZuw",
//		"bill_number": "12345678901234567890",
//		"mid": "008800223497"
//	}
//
// example response
//
//	{
//		"code": "00",
//		"message": "success",
//		"request_id": "XwVjF5zfuHhrDZuw",
//		"customer_pan": "9360001110000000019",
//		"amount": "10000.00",
//		"transaction_datetime": "2021-02-25T13:36:13",
//		"amount_fee": "1000.00",
//		"rrn": "123456789012",
//		"bill_number": "12345678901234567890",
//		"issuer_code": "93600013",
//		"customer_name": "John Doe",
//		"terminal_id": "00005771",
//		"merchant_id": "008800223497",
//		"stan": "210226",
//		"merchant_name": "Sukses Makmur Bendungan Hilir",
//		"approval_code": "00",
//		"merchant_pan": "936000131600000003",
//		"mcc": "5814",
//		"merchant_city": "Jakarta Pusat",
//		"merchant_country": "ID",
//		"currency_code": "360",
//		"payment_status": "00",
//		"payment_description": "Payment Success"
//	}
func (bni Bni) CheckStatusTransaction(r IRequest) (M, error) {
	accessToken, err := bni.getAccessToken()
	if err != nil {
		return nil, err
	}

	bodyMap := r.Payload()
	bodyMap.SetValueIfEmpty("mid", bni.MerchantID)
	requestId := bodyMap.GetValue("request_id")
	billNumber := bodyMap.GetValue("bill_number")
	message := fmt.Sprintf("%s:%s:%s", requestId, bni.MerchantID, billNumber)
	headers := M{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + accessToken,
		"X-Signature":   bni.getSignature(message),
	}
	uri := fmt.Sprintf("%s/%s", bni.Host, "check-status/inquiry")
	data := makeRequest(uri, bodyMap, headers)
	return data.Data, data.Err
}

type BniGenerateQRCodeRequest struct {
	RequestID string
	Amount    string
	QRExpired string
}

func (r BniGenerateQRCodeRequest) Payload() M {
	return M{
		"request_id": r.RequestID,
		"amount":     r.Amount,
		"qr_expired": r.QRExpired,
	}
}

type BniCheckStatusTransactionRequest struct {
	RequestID  string
	BillNumber string
}

func (r BniCheckStatusTransactionRequest) Payload() M {
	return M{
		"request_id":  r.RequestID,
		"bill_number": r.BillNumber,
	}
}
