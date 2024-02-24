package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		fmt.Println("Getting token from web")
		tok = getTokenFromWeb(config)
		fmt.Println("got token from web, Saving")
		saveToken(tokFile, tok)
		fmt.Println("saved token")
	}
	fmt.Println("Returning config")
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	fmt.Println("I'm about exchange")

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	fmt.Println("I'm here")

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	fmt.Println("I'm here....")

	client := getClient(config)

	fmt.Println("I'm here 2")

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	fmt.Println("I'm here 3")

	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("No labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range r.Labels {
		fmt.Printf("- %s\n", l.Name)
	}

	result, err := srv.Users.Messages.List(user).Q("-label:scheduled").MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve message list: %v", err)
	}

	fmt.Println("Labels:")
	for _, l := range result.Messages {
		fmt.Printf("- %s\n", l.Id)
		GetEmail(srv, l.Id)
		break
	}

}

func GetEmail(srv *gmail.Service, messageId string) {

	gmailMessageResposne, err := srv.Users.Messages.Get("me", messageId).Format("RAW").Do()
	if err != nil {
		log.Println("error when getting mail content: ", err)
	}

	if gmailMessageResposne != nil {

		decodedData, err := base64.RawURLEncoding.DecodeString(gmailMessageResposne.Raw)

		if err != nil {
			log.Println("error b64 decoding: ", err)
		}
		fmt.Printf("- %s\n", decodedData)
	}
}
