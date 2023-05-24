package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
)

const (
	BTCRateAPI = "https://api.coingecko.com/api/v3/coins/bitcoin?localization=false&tickers=false&market_data=true&community_data=false&developer_data=false&sparkline=false"
)

const emailsDBPath = "./data/emails.json"

type CoinGeckoResponse struct {
	MarketData struct {
		CurrentPrice struct {
			UAH float64 `json:"uah"`
		} `json:"current_price"`
	} `json:"market_data"`
}

type BitcoinRateUAH struct {
	PriceUAH float64 `json:"price_uah"`
}

type EmailList struct {
	Emails []string `json:"emails"`
}

var emailList EmailList

func main() {
	loadEmails()

	http.HandleFunc("/api/rate", getBTCRate)
	http.HandleFunc("/api/subscribe", subscribeEmail)
	http.HandleFunc("/api/sendEmails", sendEmails)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func parseBTCRate() (CoinGeckoResponse, error) {
	var result CoinGeckoResponse
	resp, err := http.Get(BTCRateAPI)
	if err != nil {
		return result, err
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err
	}
	err = resp.Body.Close()
	if err != nil {
		return result, err
	}
	return result, nil
}

func getBTCRate(w http.ResponseWriter, r *http.Request) {
	result, err := parseBTCRate()
	uahRate := BitcoinRateUAH{PriceUAH: result.MarketData.CurrentPrice.UAH}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(uahRate)
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func subscribeEmail(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")

	for _, e := range emailList.Emails {
		if e == email {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
	}

	emailList.Emails = append(emailList.Emails, email)
	saveEmails()

	_, err := fmt.Fprintf(w, "Email %s added successfully", email)
	if err != nil {
		return
	}
}

func sendEmails(w http.ResponseWriter, r *http.Request) {
	result, err := parseBTCRate()

	msg := fmt.Sprintf(
		"Subject: BTC Rate Update\n\nCurrent BTC to UAH rate: %f",
		result.MarketData.CurrentPrice.UAH,
	)
	for _, email := range emailList.Emails {
		sendEmail(email, msg)
	}

	_, err = fmt.Fprintf(w, "Emails sent successfully")
	if err != nil {
		return
	}
}

func loadEmails() {
	// Check if the file exists
	_, err := os.Stat(emailsDBPath)
	if os.IsNotExist(err) {
		emailList = EmailList{Emails: []string{}}
		return
	}

	file, _ := os.ReadFile(emailsDBPath)
	err = json.Unmarshal(file, &emailList)
	if err != nil {
		return
	}
}

func saveEmails() {
	file, _ := json.MarshalIndent(emailList, "", " ")
	err := os.WriteFile(emailsDBPath, file, 0644)
	if err != nil {
		return
	}
}

func sendEmail(to string, msg string) {
	from := os.Getenv("GMAIL_EMAIL")
	pass := os.Getenv("GMAIL_PASSWORD")

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Print("Email sent")
}
