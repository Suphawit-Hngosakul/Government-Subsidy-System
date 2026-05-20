package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"government-subsidy-system/backend/domain"
)

const geminiEndpoint = "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent"

const idCardPrompt = `จากรูปบัตรประชาชนนี้ กรุณาอ่านข้อมูลและตอบกลับเป็น JSON เท่านั้น ไม่ต้องมีข้อความอื่น:
{
  "nationalId": "เลขบัตรประชาชน 13 หลัก ตัวเลขล้วนไม่มีเครื่องหมาย",
  "fullName": "ชื่อ-นามสกุลภาษาไทย",
  "dateOfBirth": "วว/ดด/ปปปป พ.ศ.",
  "laserCode": "รหัส laser ด้านหลังบัตร เช่น AA1-1234567-89 ถ้าไม่เห็นใส่ string ว่าง",
  "address": "ที่อยู่ตามบัตร"
}
ถ้าไม่เห็นข้อมูลใดให้ใส่ string ว่าง ตอบเป็น JSON เท่านั้น`

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callGeminiVision(apiKey string, imageBytes []byte, mimeType string) (*domain.OCRResult, error) {
	if mimeType == "" {
		mimeType = http.DetectContentType(imageBytes)
	}

	reqBody := geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{
				{InlineData: &geminiInlineData{
					MimeType: mimeType,
					Data:     base64.StdEncoding.EncodeToString(imageBytes),
				}},
				{Text: idCardPrompt},
			},
		}},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini: http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty response")
	}

	raw := geminiResp.Candidates[0].Content.Parts[0].Text
	jsonStr := extractJSONBlock(raw)

	var parsed struct {
		NationalID  string `json:"nationalId"`
		FullName    string `json:"fullName"`
		DateOfBirth string `json:"dateOfBirth"`
		LaserCode   string `json:"laserCode"`
		Address     string `json:"address"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("gemini: parse JSON block: %w", err)
	}

	return &domain.OCRResult{
		NationalID:  parsed.NationalID,
		FullName:    parsed.FullName,
		DateOfBirth: parsed.DateOfBirth,
		LaserCode:   parsed.LaserCode,
		Address:     parsed.Address,
	}, nil
}

// extractJSONBlock strips markdown code fences and finds the first {...} block.
func extractJSONBlock(text string) string {
	text = strings.TrimSpace(text)
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")
	if m := re.FindStringSubmatch(text); len(m) > 1 {
		return m[1]
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}
