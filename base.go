package main

import (
	//"context"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"context"

	//"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	//"golang.org/x/oauth2/clientcredentials"
)

const (
	redirectURI = "http://localhost:8001/callback"
	state       = "xyzzy"
)

// UserConfig stores configuration for each user
type UserConfig struct {
	ClientID     string
	ClientSecret string
	AccessToken string
	RefreshToken string
	TokenExpiry time.Time
}

// userConfigs mao stores user configurations based on a unique user index
var userConfigs = map[string]UserConfig{
	"user1": {
		ClientID:     "b4a18d879ebf4f58ae75a32e56d3f9ac",
		ClientSecret: "54bef8595fb2420784389cfa98e4ae89",
	},
	"user2": {
		ClientID:     "65ed0f5963304995980f00ecfcdf2bb2",
		ClientSecret: "87d2e93863624c019a9ea0ab31844ae8",
	},
}

var (
	oauthConfigs = make(map[string]*oauth2.Config)
)

func init() {
	//Initialize OAuth configurations for each user
	for user, config := range userConfigs {
		oauthConfigs[user] = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  redirectURI + "?user=" + user,
			Scopes:       []string{"user-library-read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.spotify.com/authorize",
				TokenURL: "https://accounts.spotify.com/api/token",
			},
		}
	}
}

// var (
// 	oauthConfig = &oauth2.Config{
// 		ClientID:     clientID,
// 		ClientSecret: clientSecret,
// 		RedirectURL:  redirectURI,
// 		Scopes:       []string{"user-library-read"},
// 		Endpoint: oauth2.Endpoint{
// 			AuthURL:  "https://accounts.spotify.com/authorize",
// 			TokenURL: "https://accounts.spotify.com/api/token",
// 		},
// 	}
// )

func loginHandler(w http.ResponseWriter, r *http.Request) {
	user := extractUserFromRequest(r)
	oauthConfig, ok := oauthConfigs[user]
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
	//http.Redirect(w, r, auth.AuthURL(state), http.StatusFound)
}

func getUserLikedSongs(token *oauth2.Token) (map[string][]string, error) {
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopeUserLibraryRead)
	client := auth.NewClient(token)

	// Map to store songs grouped by artist
	songsByArtist := make(map[string][]string)

	// Set the limit for the number of tracks per page
	limit := 50

	// Retrieve the first page of liked songs
	tracks, err := client.CurrentUsersTracksOpt(&spotify.Options{Limit: &limit})
	if err != nil {
		return nil, err
	}

	// Iterate through the pages
	for {
		// Iterate through tracks and group by artist
		for _, item := range tracks.Tracks {
			artists := make([]string, len(item.FullTrack.Artists))
			for i, artist := range item.FullTrack.Artists {
				artists[i] = artist.Name
			}

			// Combine artists and track name for display
			artistAndTrack := fmt.Sprintf("%s - %s", strings.Join(artists, ", "), item.FullTrack.Name)

			// Check if the artistAndTrack already exists in the map
			if _, exists := songsByArtist[artistAndTrack]; !exists {
				// If not, initialize a new slice for that artistAndTrack
				songsByArtist[artistAndTrack] = make([]string, 0)
			}

			// Append to the slice
			songsByArtist[artistAndTrack] = append(songsByArtist[artistAndTrack], artistAndTrack)
		}

		// Check if there are more pages
		if tracks.Next == "" {
			break
		}

		// Retrieve the next page using the URL in the Next field
		err := client.NextPage(tracks)
		if err != nil {
			return nil, err
		}
	}

	return songsByArtist, nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	user := extractUserFromRequest(r)
	oauthConfig, ok := oauthConfigs[user]
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
		return
	}

	//store the obtained token in UserConfig
	userConfig := userConfigs[user]
	userConfig.AccessToken = token.AccessToken
	userConfig.RefreshToken= token.RefreshToken
	userConfig.TokenExpiry= token.Expiry
	userConfigs[user]= userConfig

	userLikedSongs, err := getUserLikedSongs(token)
	if err != nil {
		http.Error(w, "Failed to get user liked songs", http.StatusInternalServerError)
		log.Fatal(err)
	}

	err = exportToPDF(userLikedSongs, user)
	if err != nil {
		http.Error(w, "Failed to export to PDF", http.StatusInternalServerError)
		log.Fatal(err)
	}

	fmt.Fprintf(w, "User Liked Songs exported to PDF for %s:\n", user)
	for artistAndTrack, _ := range userLikedSongs {
		fmt.Fprintf(w, "- %s\n", artistAndTrack)
	}
}

func getUser2LikedSongs(token *oauth2.Token) (map[string][]string, error) {
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopeUserLibraryRead)
	client := auth.NewClient(token)

	// Map to store songs grouped by artist
	songsByArtist := make(map[string][]string)

	// Set the limit for the number of tracks per page
	limit := 50

	// Retrieve the first page of liked songs
	tracks, err := client.CurrentUsersTracksOpt(&spotify.Options{Limit: &limit})
	if err != nil {
		return nil, err
	}

	// Iterate through the pages
	for {
		// Iterate through tracks and group by artist
		for _, item := range tracks.Tracks {
			artists := make([]string, len(item.FullTrack.Artists))
			for i, artist := range item.FullTrack.Artists {
				artists[i] = artist.Name
			}

			// Combine artists and track name for display
			artistAndTrack := fmt.Sprintf("%s - %s", strings.Join(artists, ", "), item.FullTrack.Name)

			// Check if the artistAndTrack already exists in the map
			if _, exists := songsByArtist[artistAndTrack]; !exists {
				// If not, initialize a new slice for that artistAndTrack
				songsByArtist[artistAndTrack] = make([]string, 0)
			}

			// Append to the slice
			songsByArtist[artistAndTrack] = append(songsByArtist[artistAndTrack], artistAndTrack)
		}

		// Check if there are more pages
		if tracks.Next == "" {
			break
		}

		// Retrieve the next page using the URL in the Next field
		err := client.NextPage(tracks)
		if err != nil {
			return nil, err
		}
	}

	return songsByArtist, nil
}

func getUser2LikedSongsHandler(w http.ResponseWriter, r *http.Request) {
	user := extractUserFromRequest(r)
	oauthConfig, ok := oauthConfigs[user]
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Couldn,t get token", http.StatusForbidden)
		log.Fatal(err)
	}

	user2LikedSongs, err := getUser2LikedSongs(token)
	if err != nil {
		http.Error(w, "Failed to get user 2 liked songs", http.StatusInternalServerError)
		log.Fatal(err)
	}
	fmt.Fprintf(w, "User 2 Liked Songs:\n")
	for artistAndTrack, _ := range user2LikedSongs {
		fmt.Fprintf(w, "-s\n", artistAndTrack)
	}
}

func exportToPDF(songsByArtist map[string][]string,user string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	//Font
	pdf.SetFont("Arial", "B", 16)

	//Add Title
	pdf.Cell(40, 10, "User Liked Songs")

	//Set font for songlist
	pdf.SetFont("Arial", "", 12)

	//add songs to pdf
	for artistAndTrack, _ := range songsByArtist {
		pdf.Cell(40, 10, fmt.Sprintf("- %s", artistAndTrack))
		pdf.Ln(-1)
	}
	err := pdf.OutputFileAndClose("user_liked_songs.pdf")
	if err != nil {
		return err
	}
	return nil
}

func compareUsersHandler(w http.ResponseWriter, r *http.Request) {
	user1 := "user1"
	user2 := "user2"

	//Retrieve token
	token1, err := getUserToken(user1)
	if err != nil{
		http.Error(w, "Could not get user token", http.StatusInternalServerError)
		return
	}

	token2, err:= getUserToken(user2)
	if err!= nil{
		http.Error(w, "Couldnt get user 2 token", http.StatusInternalServerError)
		return
	}
	likedSongs1, err:= getUserLikedSongs(token1)
	if err!= nil{
		http.Error(w, "Failed to get user 1 liked songs", http.StatusInternalServerError)
		return
	}
	likedSongs2, err := getUser2LikedSongs(token2)
	if err != nil{
		http.Error(w, "Failed to get user 2 liked songs", http.StatusInternalServerError)
		return
	}
	
	//Compare liked songs
	commonLikedSongs := make(map[string]struct{})
	for song := range likedSongs1 {
		if _, exists := likedSongs2[song]; exists {
			commonLikedSongs[song] = struct{}{}
		}
	}

	//Return common liked songs in the response
	fmt.Fprintf(w, "Common liked songs between User1 and User 2:\n")
	for song := range commonLikedSongs {
		fmt.Fprintf(w, "-%s\n", song)
	}
}

func getUserToken(user string) (*oauth2.Token, error){
	userConfig, ok := userConfigs[user]
	if !ok {
		return nil, fmt.Errorf("Invalid user :%s", user)
	}
	// Check if the token is still valid; if not, refresh the token
    if time.Now().After(userConfig.TokenExpiry) {
        tokenSource := oauthConfigs[user].TokenSource(context.Background(), &oauth2.Token{
            RefreshToken: userConfig.RefreshToken,
        })

        // Get the refreshed token
        token, err := tokenSource.Token()
        if err != nil {
            return nil, fmt.Errorf("Failed to refresh token for user %s: %v", user, err)
        }

        // Update the stored token with the refreshed token
        userConfig.AccessToken = token.AccessToken
        userConfig.RefreshToken = token.RefreshToken
        userConfig.TokenExpiry = token.Expiry
        userConfigs[user] = userConfig
    }

    return &oauth2.Token{
        AccessToken:  userConfig.AccessToken,
        RefreshToken: userConfig.RefreshToken,
        Expiry:       userConfig.TokenExpiry,
    }, nil
}

func extractUserFromRequest(r *http.Request) string{
	return r.URL.Query().Get("user")
}
func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/user2LikedSongs", getUser2LikedSongsHandler)
	http.HandleFunc("/compareUsers", compareUsersHandler)
	fmt.Println("Server is starting on :8001...")
	log.Fatal(http.ListenAndServe(":8001", nil))
}
