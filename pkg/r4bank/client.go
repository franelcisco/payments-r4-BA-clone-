package r4bank

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type RestClient struct {
	baseURL string
	token   string
	client  *http.Client
	logger  *zap.Logger
}

func NewClient(
	endpoint string,
	token string,
	logger *zap.Logger,
) *RestClient {
	uuidToken := GenerateAuthToken(token, "boneappetitR4ServiceSecretKey")
	if !ValidateAuthToken(token, "boneappetitR4ServiceSecretKey", uuidToken) {
		logger.Error("Invalid R4Bank token")
		return nil
	} else {
		logger.Info("R4Bank token is valid", zap.String("uuidToken", uuidToken))
	}
	return &RestClient{
		baseURL: endpoint,
		token:   token,
		client:  &http.Client{Timeout: 20 * time.Second},
		logger:  logger,
	}
}

func GenerateAuthToken(key, message string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidateAuthToken(key, message, token string) bool {
	expected := GenerateAuthToken(key, message)
	return hmac.Equal([]byte(expected), []byte(token))
}

func (r *RestClient) Do(
	ctx context.Context,
	hmacInput string,
	payload map[string]string,
	endpoint string,
) ([]byte, error) {
	mac := hmac.New(sha256.New, []byte(r.token))
	mac.Write([]byte(hmacInput))
	auth := hex.EncodeToString(mac.Sum(nil))

	body, err := json.Marshal(payload)
	if err != nil {
		r.logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	url := fmt.Sprintf("%s/%s", r.baseURL, endpoint)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url, bytes.NewReader(body),
	)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth)
	req.Header.Set("Commerce", r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Error(err.Error(), zap.Any("payload", payload))
		return nil, fmt.Errorf("error en request: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r.logger.Error("R4 API error: ", zap.String("body", string(data)), zap.Any("payload", payload))
		return nil, fmt.Errorf("R4 API error: %s", string(data))
	}

	return data, nil
}
