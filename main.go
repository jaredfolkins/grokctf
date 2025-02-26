package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var (
	tpl   *template.Template
	users []*User
)

type User struct {
	ID       int
	Username string
	Password string
}

type LoginData struct {
	Username string
	Password string
}

func main() {

	users = []*User{
		{ID: 1, Username: "jared", Password: "lovesevan"},
		{ID: 2, Username: "evan", Password: "lovesjared"},
		{ID: 3, Username: "grok", Password: "hacking"},
	}

	// Set up logging
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Parse templates
	tpl = template.Must(template.ParseFiles("login.html"))

	// Set up routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/userquery", userQueryHandler) // Renamed from /idor

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Redirecting %s to /login?username=&password=", r.RemoteAddr)
	http.Redirect(w, r, "/login?username=&password=", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var username, password string

	// Parse form data for POST requests
	if r.Method == "POST" {
		r.ParseForm()
		username = r.FormValue("username")
		password = r.FormValue("password")
	} else {
		username = r.URL.Query().Get("username")
		password = r.URL.Query().Get("password")
	}

	// Handle login attempts (both GET and POST)
	if (r.Method == "GET" || r.Method == "POST") && username != "" && password != "" {
		log.Printf("Login attempt from %s with username: %s via %s", r.RemoteAddr, username, r.Method)

		for _, user := range users {
			if username == user.Username && password == user.Password {
				log.Printf("Successful login from %s for user: %s", r.RemoteAddr, user.Username)
				w.Write([]byte(fmt.Sprintf("Login successful for %s! The flag url is https://x.com/jf0lkins", user.Username)))
				return
			}
		}

		w.Write([]byte("login failed!"))
		return
	}

	// Serve login page for GET requests without complete credentials
	if r.Method == "GET" {
		log.Printf("Serving login page to %s", r.RemoteAddr)
		tpl.ExecuteTemplate(w, "login.html", LoginData{Username: username, Password: password})
		return
	}

	// If POST without credentials, return error
	if r.Method == "POST" {
		log.Printf("Invalid POST login attempt from %s: missing credentials", r.RemoteAddr)
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}
}

func userQueryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form data for POST requests
	if r.Method == "POST" {
		r.ParseForm()
	}

	// Get userid from either query string (GET) or form data (POST)
	userIDStr := r.FormValue("userid")
	if userIDStr == "" && r.Method == "GET" {
		userIDStr = r.URL.Query().Get("userid")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Invalid userid from %s: %s", r.RemoteAddr, userIDStr)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("User query request from %s for userid: %d via %s", r.RemoteAddr, userID, r.Method)

	targetUser := &User{}

	for _, user := range users {
		if user.ID == userID {
			targetUser = user
			break
		}
	}

	if targetUser.ID == 0 {
		log.Printf("User not found for userid: %d", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Return the credentials
	fmt.Fprintf(w, "Username: %s\nPassword: %s", targetUser.Username, targetUser.Password)
	log.Printf("Returned credentials to %s for userid: %d (username: %s password %s)", r.RemoteAddr, userID, targetUser.Username, targetUser.Password)
}
