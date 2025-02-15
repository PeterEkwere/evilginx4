package core

import (
    "bytes"
    "encoding/json"
    "fmt"
    "mime/multipart"
    "net/http"
)

type TelegramNotifier struct {
    botToken string
    chatID   string
}

func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
    return &TelegramNotifier{
        botToken: botToken,
        chatID:   chatID,
    }
}

func (t *TelegramNotifier) SendMessage(message string) error {
	log.Info("[Telegram] Attempting to send message") 
    if t.botToken == "" || t.chatID == "" {
		log.Info("[Telegram] nil is being called") 
        return nil // silently ignore if telegram is not configured
    }

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
    
    payload := map[string]string{
        "chat_id": t.chatID,
        "text":    message,
        "parse_mode": "HTML",
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (t *TelegramNotifier) SendFile(filename string, content string) error {
    if t.botToken == "" || t.chatID == "" {
        return nil // silently ignore if telegram is not configured
    }

    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", t.botToken)
    
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    writer.WriteField("chat_id", t.chatID)
    
    part, err := writer.CreateFormFile("document", filename)
    if err != nil {
        return err
    }
    part.Write([]byte(content))
    
    writer.Close()
    
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}