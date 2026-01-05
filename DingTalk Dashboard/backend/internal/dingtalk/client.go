package dingtalk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	tokenURL          = "https://oapi.dingtalk.com/gettoken"
	approvalListURL   = "https://oapi.dingtalk.com/topapi/processinstance/listids"
	approvalDetailURL = "https://oapi.dingtalk.com/topapi/processinstance/get"
	userInfoURL       = "https://oapi.dingtalk.com/topapi/v2/user/get"
)

// Client is a DingTalk API client
type Client struct {
	appKey      string
	appSecret   string
	accessToken string
	tokenExpiry time.Time
	mu          sync.RWMutex
	httpClient  *http.Client
}

// NewClient creates a new DingTalk client
func NewClient(appKey, appSecret string) *Client {
	return &Client{
		appKey:    appKey,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// getAccessToken gets or refreshes the access token
func (c *Client) getAccessToken() (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	// Fetch new token
	reqURL := fmt.Sprintf("%s?appkey=%s&appsecret=%s", tokenURL, c.appKey, c.appSecret)
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("DingTalk API error: %s", result.ErrMsg)
	}

	c.accessToken = result.AccessToken
	// Set expiry 5 minutes before actual expiry for safety
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	return c.accessToken, nil
}

// GetApprovalInstanceIDs gets list of approval instance IDs (only startTime, no endTime)
func (c *Client) GetApprovalInstanceIDs(processCode string, startTime time.Time, cursor int64, size int) (*ApprovalListResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s?access_token=%s", approvalListURL, token)

	data := url.Values{}
	data.Set("process_code", processCode)
	data.Set("start_time", fmt.Sprintf("%d", startTime.UnixMilli()))
	// Note: end_time is intentionally not set as per requirements
	data.Set("cursor", fmt.Sprintf("%d", cursor))
	data.Set("size", fmt.Sprintf("%d", size))

	resp, err := c.httpClient.Post(reqURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to get instance IDs: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ApprovalListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("DingTalk API error: %s", result.ErrMsg)
	}

	return &result, nil
}

// GetApprovalInstanceDetail gets detailed info for an instance
func (c *Client) GetApprovalInstanceDetail(processInstanceID string) (*ApprovalDetailResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s?access_token=%s", approvalDetailURL, token)

	data := url.Values{}
	data.Set("process_instance_id", processInstanceID)

	resp, err := c.httpClient.Post(reqURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to get instance detail: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ApprovalDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("DingTalk API error: %s", result.ErrMsg)
	}

	return &result, nil
}

// GetUserInfo gets user information by user ID
func (c *Client) GetUserInfo(userID string) (*UserInfoResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s?access_token=%s", userInfoURL, token)

	reqBody := fmt.Sprintf(`{"userid":"%s"}`, userID)

	resp, err := c.httpClient.Post(reqURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result UserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("DingTalk API error: %s", result.ErrMsg)
	}

	return &result, nil
}

// GetUserName gets user name by user ID with caching
func (c *Client) GetUserName(userID string, cache map[string]string) string {
	if name, ok := cache[userID]; ok {
		return name
	}

	info, err := c.GetUserInfo(userID)
	if err != nil {
		return userID // Return ID if can't get name
	}

	name := info.Result.Name
	cache[userID] = name
	return name
}
