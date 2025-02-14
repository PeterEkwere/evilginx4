package database

import (
    "bytes"
    "encoding/json"
    "fmt"
    "mime/multipart"  // Changed from just "multipart"
    "net/http"
)

type TelegramConfig struct {
    BotToken string
    ChatID   string
}

func sendTelegramMessage(config TelegramConfig, message string) error {
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
    
    payload := map[string]string{
        "chat_id": config.ChatID,
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

func sendTelegramFile(config TelegramConfig, filename string, content string) error {
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", config.BotToken)
    
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Add chat_id field
    writer.WriteField("chat_id", config.ChatID)
    
    // Create document part
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