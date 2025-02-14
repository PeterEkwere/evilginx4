package database

import (
    "bytes"
    "encoding/json"
    "fmt"
    "mime/multipart"
    "net/http"
    "log" // Add this for debugging
)

type TelegramConfig struct {
    BotToken string
    ChatID   string
}

func sendTelegramMessage(config TelegramConfig, message string) error {
    log.Printf("Attempting to send Telegram message. Bot Token: %s, Chat ID: %s", config.BotToken, config.ChatID)
    
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
    
    payload := map[string]string{
        "chat_id": config.ChatID,
        "text":    message,
        "parse_mode": "HTML",
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshaling JSON: %v", err)
        return err
    }

    log.Printf("Sending POST request to: %s", url)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error making POST request: %v", err)
        return err
    }
    defer resp.Body.Close()

    // Read and log response
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    log.Printf("Telegram API Response: %v", result)

    return nil
}

func sendTelegramFile(config TelegramConfig, filename string, content string) error {
    log.Printf("Attempting to send Telegram file: %s", filename)
    
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", config.BotToken)
    
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Add chat_id field
    writer.WriteField("chat_id", config.ChatID)
    
    // Create document part
    part, err := writer.CreateFormFile("document", filename)
    if err != nil {
        log.Printf("Error creating form file: %v", err)
        return err
    }
    part.Write([]byte(content))
    
    writer.Close()
    
    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        log.Printf("Error creating request: %v", err)
        return err
    }
    
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending file: %v", err)
        return err
    }
    defer resp.Body.Close()

    // Read and log response
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    log.Printf("Telegram API Response for file: %v", result)
    
    return nil
}