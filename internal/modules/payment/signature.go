package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type SignatureVerifier struct {
	webhookSecret string
	keySecret     string
}

func NewSignatureVerifier(
	keySecret string,
	webhookSecret string,
) *SignatureVerifier {

	return &SignatureVerifier{
		keySecret:     keySecret,
		webhookSecret: webhookSecret,
	}
}

//
// Checkout Verification
//
// Razorpay sends:
//
// razorpay_order_id
// razorpay_payment_id
// razorpay_signature
//
// Signature = HMAC_SHA256(
//      order_id + "|" + payment_id,
//      key_secret
// )
//

func (s *SignatureVerifier) VerifyCheckoutSignature(
	orderID string,
	paymentID string,
	signature string,
) error {

	payload := orderID + "|" + paymentID

	hash := hmac.New(
		sha256.New,
		[]byte(s.keySecret),
	)

	hash.Write([]byte(payload))

	expected := hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal(
		[]byte(expected),
		[]byte(signature),
	) {
		return ErrInvalidSignature
	}

	return nil
}

//
// Webhook Verification
//
// Razorpay signs the entire request body.
//
// Signature =
//
// HMAC_SHA256(
//      request_body,
//      webhook_secret
// )
//

func (s *SignatureVerifier) VerifyWebhookSignature(
	body []byte,
	signature string,
) error {

	hash := hmac.New(
		sha256.New,
		[]byte(s.webhookSecret),
	)

	hash.Write(body)

	expected := hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal(
		[]byte(expected),
		[]byte(signature),
	) {
		return ErrInvalidSignature
	}

	return nil
}

//
// Generic helper
//

func VerifyHMACSHA256(
	payload []byte,
	secret string,
	signature string,
) error {

	hash := hmac.New(
		sha256.New,
		[]byte(secret),
	)

	hash.Write(payload)

	expected := hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal(
		[]byte(expected),
		[]byte(signature),
	) {
		return errors.New("signature verification failed")
	}

	return nil
}