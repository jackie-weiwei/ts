package ts

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AppleLoginToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
}

// create client_secret
func getAppleSecret(secret, keyId, teamId, bundleId string) string {
	token := &jwt.Token{
		Header: map[string]interface{}{
			"alg": "ES256",
			"kid": keyId,
		},
		Claims: jwt.MapClaims{
			"iss": teamId,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"aud": "https://appleid.apple.com",
			"sub": bundleId,
		},
		Method: jwt.SigningMethodES256,
	}

	ecdsaKey, _ := authKeyFromBytes([]byte(secret))
	ss, _ := token.SignedString(ecdsaKey)
	return ss
}

// create private key for jwt sign
func authKeyFromBytes(key []byte) (*ecdsa.PrivateKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, errors.New("token: AuthKey must be a valid .p8 PEM file")
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		return nil, err
	}

	var pkey *ecdsa.PrivateKey
	var ok bool
	if pkey, ok = parsedKey.(*ecdsa.PrivateKey); !ok {
		return nil, errors.New("token: AuthKey must be of type ecdsa.PrivateKey")
	}

	return pkey, nil
}

// do http request
func httpRequest(method, addr string, params map[string]string) ([]byte, int, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	var request *http.Request
	var err error
	if request, err = http.NewRequest(method, addr, strings.NewReader(form.Encode())); err != nil {
		return nil, 0, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var response *http.Response
	if response, err = http.DefaultClient.Do(request); nil != err {
		return nil, 0, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, 0, err
	}
	return data, response.StatusCode, nil
}

func AppleLogin(authorizationCode, certSecret, keyID, teamID, bundleId, redirectUrl string) string {
	data, _, err := httpRequest("POST", "https://appleid.apple.com/auth/token", map[string]string{
		"client_id":     bundleId,
		"client_secret": getAppleSecret(certSecret, keyID, teamID, bundleId),
		"code":          authorizationCode,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectUrl,
	})

	if err != nil {
		return ""
	}

	var tt AppleLoginToken
	err = json.Unmarshal(data, &tt)
	if err != nil {
		return ""
	}

	t, err := verifyIDToken(tt.IdToken)

	if err != nil {
		fmt.Println("Token verification failed:", err)
		return ""
	}

	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {

		fmt.Println("Token verified. Claims:", claims)
		return claims["email"].(string)
	} else {
		fmt.Println("Invalid token")
		return ""
	}
}

// 定义Apple公钥的结构体
type applePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type applePublicKeys struct {
	Keys []applePublicKey `json:"keys"`
}

// 从Apple的JWKs端点获取公钥
func getApplePublicKeys() (*applePublicKeys, error) {
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var keys applePublicKeys
	err = json.Unmarshal(body, &keys)
	if err != nil {
		return nil, err
	}

	return &keys, nil
}

// 根据kid找到匹配的公钥
func findMatchingKey(kid string, keys *applePublicKeys) (*rsa.PublicKey, error) {
	for _, key := range keys.Keys {
		if key.Kid == kid {
			n, err := base64URLDecodeToBigInt(key.N)
			if err != nil {
				return nil, err
			}
			e, err := base64URLDecodeToBigInt(key.E)
			if err != nil {
				return nil, err
			}
			return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
		}
	}
	return nil, fmt.Errorf("matching key not found")
}

// Base64 URL 解码到大整数
func base64URLDecodeToBigInt(base64url string) (*big.Int, error) {
	// Base64 URL 解码
	decoded, err := jwt.DecodeSegment(base64url)
	if err != nil {
		return nil, err
	}

	// 转换为大整数
	n := big.NewInt(0)
	n.SetBytes(decoded)
	return n, nil
}

// 验证ID Token
func verifyIDToken(tokenString string) (*jwt.Token, error) {
	// 分割JWT，获取Header部分
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected 3 parts but got %d", len(parts))
	}

	// Base64 URL 解码Header
	headerJSON, err := jwt.DecodeSegment(parts[0])
	if err != nil {
		return nil, err
	}

	// 解析Header为map
	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, err
	}

	// 从Header中获取kid
	kid, ok := header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid not found in token header")
	}

	// 获取Apple的公钥
	keys, err := getApplePublicKeys()
	if err != nil {
		return nil, err
	}

	// 根据kid找到匹配的公钥
	publicKey, err := findMatchingKey(kid, keys)
	if err != nil {
		return nil, err
	}

	// 使用找到的公钥解析并验证JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保token使用的是RSA签名
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
