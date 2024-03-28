package utils

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func IsLoginTaken(login string, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE login = $1",
		login).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func TokenExpiresTime(expires string) (time.Time, error) {
	index := strings.Index(expires, "d")
	days, err := strconv.Atoi(expires[:index])
	if err != nil {
		return time.Now(), err
	}
	expirationTime := time.Now().Add(time.Duration(days*24) * time.Hour)
	return expirationTime, nil
}

func getImageSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	size := resp.Header.Get("Content-Length")
	return stringToSize(size)
}

func stringToSize(size string) (int64, error) {
	var sz int64
	_, err := fmt.Sscan(size, &sz)
	return sz, err
}

func IsImageURLWithSizeLimit(url string, maxSize int64) (bool, error) {
	imageExtensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".svg"}
	isImage := false
	for _, ext := range imageExtensions {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			isImage = true
			break
		}
	}
	if !isImage {
		return false, nil
	}

	size, err := getImageSize(url)
	if err != nil {
		return false, err
	}

	if size <= maxSize {
		return true, nil
	}
	return false, nil
}

func GetUserLoginByID(db *sql.DB, userID int) (string, error) {
	var userLogin string
	query := "SELECT login FROM users WHERE user_id = $1"
	err := db.QueryRow(query, userID).Scan(&userLogin)
	if err != nil {
		return "", err
	}
	return userLogin, nil
}
