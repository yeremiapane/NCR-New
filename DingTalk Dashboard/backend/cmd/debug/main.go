package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	tokenURL        = "https://oapi.dingtalk.com/gettoken"
	approvalListURL = "https://oapi.dingtalk.com/topapi/processinstance/listids"
)

type NCRApproval struct {
	ID                string `gorm:"column:id"`
	ProcessInstanceID string `gorm:"column:process_instance_id"`
	BusinessID        string `gorm:"column:business_id"`
	Title             string `gorm:"column:title"`
	Status            string `gorm:"column:status"`
}

func (NCRApproval) TableName() string {
	return "ncr_approvals"
}

func main() {
	godotenv.Load()
	appKey := "ding1tr6td1dwfrs6dmo"
	appSecret := "lWZ6s0rmbmXg9qxQ7kixUbKT1m68M_7O7fLI4vO7q98K7rINiYxAwUYCdOwt8PP0"
	processCode := "PROC-C4B4714A-07A7-44BF-858D-4E842C529736"
	targetInstanceID := "LuezsLqmSPO7HdVmH02MUg08541767340915"

	// Get access token
	tokenResp, _ := http.Get(fmt.Sprintf("%s?appkey=%s&appsecret=%s", tokenURL, appKey, appSecret))
	defer tokenResp.Body.Close()

	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(tokenResp.Body).Decode(&tokenResult)

	// Fetch from 5 days ago
	startTime := time.Now().AddDate(0, 0, -5)
	fmt.Printf("Start time: %s\n\n", startTime.Format("2006-01-02 15:04:05"))

	listData := url.Values{}
	listData.Set("process_code", processCode)
	listData.Set("start_time", fmt.Sprintf("%d", startTime.UnixMilli()))
	listData.Set("cursor", "0")
	listData.Set("size", "20")

	listResp, _ := http.Post(
		fmt.Sprintf("%s?access_token=%s", approvalListURL, tokenResult.AccessToken),
		"application/x-www-form-urlencoded",
		strings.NewReader(listData.Encode()),
	)
	defer listResp.Body.Close()

	listBody, _ := io.ReadAll(listResp.Body)

	var listResult struct {
		Result struct {
			List []string `json:"list"`
		} `json:"result"`
	}
	json.Unmarshal(listBody, &listResult)

	fmt.Printf("=== All %d instance IDs from API ===\n", len(listResult.Result.List))
	for i, id := range listResult.Result.List {
		marker := ""
		if id == targetInstanceID {
			marker = " <-- TARGET"
		}
		fmt.Printf("[%d] %s%s\n", i+1, id, marker)
	}

	// Connect to DB and check each one
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:allure2025@localhost:5434/ncr_dashboard?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("\nFailed to connect to DB: %v\n", err)
		return
	}

	fmt.Printf("\n=== Checking database for each instance ===\n")
	for i, id := range listResult.Result.List {
		var approval NCRApproval
		result := db.Where("process_instance_id = ?", id).First(&approval)
		status := "NOT FOUND"
		if result.Error == nil {
			status = fmt.Sprintf("FOUND (business_id=%s)", approval.BusinessID)
		}
		marker := ""
		if id == targetInstanceID {
			marker = " <-- TARGET"
		}
		fmt.Printf("[%d] %s: %s%s\n", i+1, id[:20], status, marker)
	}
}
