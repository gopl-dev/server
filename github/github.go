package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

// TODO: [REVIEW] should I move MakeHMAC and IsValidHMAC to a crypto package?
// (These functions are specifically for GitHub API requirements, so they are located here for now)

// SigPrefix
// GitHub adds the prefix "sha1=" to the HMAC hash.
const SigPrefix = "sha1="

// MakeHMAC generates an HMAC for the given body and key.
// https://developer.github.com/webhooks/event-payloads/#delivery-headers
// The X-Hub-Signature header contains the HMAC hex digest of the request body.
// This header is sent if the webhook is configured with a secret.
// The HMAC hex digest is generated using the SHA1 hash function and the secret as the HMAC key.
// Note: GitHub also adds the prefix "sha1=" to the hash.
func MakeHMAC(body, key string) (hash string, err error) {
	h := hmac.New(sha1.New, []byte(key))
	_, err = h.Write([]byte(body))
	if err != nil {
		return
	}

	hash = SigPrefix + fmt.Sprintf("%x", h.Sum(nil))
	return
}

// IsValidHMAC verifies if the given HMAC hash is valid for the provided body and key.
func IsValidHMAC(body []byte, hash, key string) (ok bool, err error) {
	h := hmac.New(sha1.New, []byte(key))
	_, err = h.Write(body)
	if err != nil {
		return
	}

	expected := fmt.Sprintf("%x", h.Sum(nil))
	ok = hash == (SigPrefix + expected)

	return
}
