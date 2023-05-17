package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type oauth1Opt struct {
	consumerKey string
	accessToken string
	signingKey  string
}

func OAuth1(consumerKey, accessToken, signingKey string) *oauth1Opt {
	return &oauth1Opt{
		consumerKey: consumerKey,
		accessToken: accessToken,
		signingKey:  signingKey,
	}
}

func (opt *oauth1Opt) Do(client defaultClient, req *http.Request) {
	parameters := Parameter{}
	parameters["oauth_consumer_key"] = opt.consumerKey
	parameters["oauth_nonce"] = "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg" // crypto.GenerateRandomAlphabet(32)
	parameters["oauth_signature_method"] = "HMAC-SHA1"
	parameters["oauth_timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	parameters["oauth_token"] = opt.accessToken
	parameters["oauth_version"] = "1.0"
	parameterString := generateParameterString(parameters, client)

	var signatureBase []string
	signatureBase = append(signatureBase, strings.ToUpper(client.method))
	signatureBase = append(signatureBase, PercentEncode(client.url))
	signatureBase = append(signatureBase, PercentEncode(parameterString))
	signatureBaseString := strings.Join(signatureBase, "&")

	h := hmac.New(sha1.New, []byte(opt.signingKey))
	h.Write([]byte(signatureBaseString))
	parameters["oauth_signature"] = base64.StdEncoding.EncodeToString(h.Sum(nil))

	var oauthParams []string
	for key, value := range parameters {
		oauthParams = append(oauthParams, PercentEncode(key)+"=\""+PercentEncode(value)+"\"")
	}

	sort.Strings(oauthParams)
	req.Header.Add("Authorization", "OAuth "+strings.Join(oauthParams, ","))
}

func generateParameterString(parameters Parameter, client defaultClient) string {
	finalParameters := Parameter{}

	for key, value := range parameters {
		finalParameters[key] = value
	}

	for key, value := range client.query {
		finalParameters[key] = value
	}

	switch body := client.body.(type) {
	// OAuth1.0 only encodes x-www-url-encoded body .
	case Parameter:
		for key, value := range body {
			finalParameters[key] = value
		}
	}

	return finalParameters.Encode()
}

type oauth2Opt struct {
	token string
}

func OAuth2(prefix, token string) *oauth2Opt {
	return &oauth2Opt{token: prefix + " " + token}
}

func (opt *oauth2Opt) Do(client defaultClient, req *http.Request) {
	req.Header.Add("Authorization", opt.token)
}
