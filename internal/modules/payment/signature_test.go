package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignatureVerifier(t *testing.T) {
	verifier := NewSignatureVerifier("checkout-secret", "webhook-secret")

	checkoutPayload := "order_1|pay_1"
	checkoutMAC := hmac.New(sha256.New, []byte("checkout-secret"))
	checkoutMAC.Write([]byte(checkoutPayload))
	checkoutSignature := hex.EncodeToString(checkoutMAC.Sum(nil))

	if err := verifier.VerifyCheckoutSignature("order_1", "pay_1", checkoutSignature); err != nil {
		t.Fatalf("VerifyCheckoutSignature() error = %v", err)
	}

	if err := verifier.VerifyCheckoutSignature("order_1", "pay_1", "bad"); err == nil {
		t.Fatal("expected invalid checkout signature to fail")
	}

	body := []byte(`{"event":"payment.captured"}`)
	webhookMAC := hmac.New(sha256.New, []byte("webhook-secret"))
	webhookMAC.Write(body)
	webhookSignature := hex.EncodeToString(webhookMAC.Sum(nil))

	if err := verifier.VerifyWebhookSignature(body, webhookSignature); err != nil {
		t.Fatalf("VerifyWebhookSignature() error = %v", err)
	}

	if err := VerifyHMACSHA256(body, "webhook-secret", webhookSignature); err != nil {
		t.Fatalf("VerifyHMACSHA256() error = %v", err)
	}
}
