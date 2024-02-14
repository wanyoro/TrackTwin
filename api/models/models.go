package models

type User struct{
	ID uint `json:"id"`
	Username string `json:"username"`
}

type Song struct {
	ID uint `json:"id"`
	Title string `json:"title"`
	Artist string `json:"artist"`
}

type LikedSong struct {
	UserID uint `json:"userId"`
	SongID uint `json:"songId"`
}

type Match struct{
	User1ID uint `json:"user1Id"`
	User2Id uint `json:"user2Id"`
}