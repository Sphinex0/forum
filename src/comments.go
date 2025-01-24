package forum

import (
	"net/http"
	"strings"
	"time"
)

// CommentsStore handles the creation of comments for a post.
func (data Page) CommentsStore(res http.ResponseWriter, req *http.Request) {
	// Ensure the request method is POST
	if req.Method != "POST" {
		data.Error(res, http.StatusMethodNotAllowed)
		return
	}

	// Extract form values
	postID := req.FormValue("post_id")
	body := req.FormValue("body")

	// Check if the user is on cooldown
	Mux.Lock()
	_, isOnCooldown := Cach[data.Id]
	Mux.Unlock()

	if isOnCooldown {
		setErrorCookie(res,"You are on cooldown! Wait for a bit and try again ^_^.", getPathFromReferer(req),  1)
		http.Redirect(res, req, "/posts/"+postID, http.StatusFound)
		return
	}

	// Validate the comment body length
	if len(body) < 1 || len(body) > 5000 {
		data.Error(res, http.StatusBadRequest)
		return
	}

	// Insert the comment into the database
	query := "INSERT INTO comments VALUES (NULL,?,?,?,?,?)"
	_, err := DB.Exec(query, data.Id, postID, body, time.Now().Unix(), time.Now().Unix())
	if err != nil {
		data.Error(res, http.StatusInternalServerError)
		return
	}

	// Add the user to the cooldown cache
	Mux.Lock()
	Cach[data.Id] = time.Now().Unix()
	Mux.Unlock()

	// Redirect the user back to the post
	http.Redirect(res, req, "/posts/"+postID, http.StatusFound)
}

// getPathFromReferer extracts the path from the Referer header.
func getPathFromReferer(req *http.Request) string {
	referer := req.Referer()
	if referer == "" {
		return "/"
	}

	// Find the first '/' after the domain and return the path
	if idx := strings.Index(referer, "://"); idx != -1 {
		if pathIdx := strings.Index(referer[idx+3:], "/"); pathIdx != -1 {
			return referer[idx+3+pathIdx:]
		}
	}
	return "/"
}
