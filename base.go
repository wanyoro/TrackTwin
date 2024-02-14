package main

import (
	//"context"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	//"strings"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"github.com/jung-kurt/gofpdf"
	//"golang.org/x/oauth2/clientcredentials"
)

const (
	clientID     = "b4a18d879ebf4f58ae75a32e56d3f9ac"
	clientSecret = "54bef8595fb2420784389cfa98e4ae89"
	redirectURI  = "http://localhost:8001/callback"
	state        = "xyzzy"
)

var (
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"user-library-read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
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
	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	// Use the access token to make requests to the Spotify API
	//client := oauthConfig.Client(r.Context(), token)
	userLikedSongs, err := getUserLikedSongs(token)
	if err != nil {
		http.Error(w, "Failed to get user liked songs", http.StatusInternalServerError)
		log.Fatal(err)
	}

	err = exportToPDF(userLikedSongs)
	if err != nil {
		http.Error(w, "Failed to export to PDF", http.StatusInternalServerError)
		log.Fatal(err)
	}
	//ch <- &client
	fmt.Fprintf(w, "User Liked Songs exported to PDF:\n")
	for artistAndTrack, songs := range userLikedSongs {
		fmt.Fprintf(w, "- %s(%d songs)\n", artistAndTrack, len(songs))
	}
}

func exportToPDF(songsByArtist map[string][]string) error {
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
		pdf.Cell(40, 10, fmt.Sprintf("- %s (%d songs)", artistAndTrack))
		pdf.Ln(-1)
	}
	err := pdf.OutputFileAndClose("user_liked_songs.pdf")
	if err!= nil{
		return err
	}
	return nil
}
func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	fmt.Println("Server is starting on :8001...")
	log.Fatal(http.ListenAndServe(":8001", nil))

}
