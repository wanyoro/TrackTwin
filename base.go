package main

import (
	//"context"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	//"strings"erer

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
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



func getUserLikedSongs(client *spotify.Client) (map[string][]string, error) {
	// Use the Spotify API client to retrieve user's liked songs
	// This example groups songs by artist for a more organized display.
	// Adjust based on the Spotify API's actual structure.
	// For Spotify, you might want to use the /v1/me/tracks endpoint
	// and extract the track names, artists, or IDs from the response.
	// Note: This is a simplified example and might not work as-is.

	// Example: Get the current user's saved tracks
	tracks, err := client.CurrentUsersTracks()
	if err != nil {
		return nil, err
	}

	// Map to store songs grouped by artist
	songsByArtist := make(map[string][]string)

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
	client := spotify.Authenticator{}.NewClient(token)
	userLikedSongs, err := getUserLikedSongs(&client)
	if err != nil {
		http.Error(w, "Failed to get user liked songs", http.StatusInternalServerError)
		log.Fatal(err)
	}
	//ch <- &client
	fmt.Fprintf(w, "User Liked Songs:\n")
	for artistAndTrack, songs := range userLikedSongs{
		fmt.Fprintf(w, "- %s(%d songs)\n", artistAndTrack, len(songs))
	}
}
func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	fmt.Println("Server is starting on :8001...")
	log.Fatal(http.ListenAndServe(":8001", nil))

	// url := auth.AuthURL(state)
	// fmt.Printf("Please log in to spotify by visiting %v\n", url)

	// select {
	// case client := <-ch:
	// 	getUserLikedSongs(client)
	// }
}
