package main

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

func PostImage(w rest.ResponseWriter, r *rest.Request) {
	authHeader := r.Header["Authorization"]
	if authHeader == nil || len(authHeader) == 0 || len(authHeader[0]) < len("Bearer ") {
		writeAuthError(w)
		return
	}

	no, ok := parseToken(authHeader[0][len("Bearer "):])
	if !ok {
		writeAuthError(w)
		return
	}

	contentType := r.Header["Content-Type"][0]

	now := strconv.FormatInt(time.Now().Unix(), 36)
	encodedNo := strconv.FormatUint(no, 36)
	ext := ""
	switch contentType {
	case "image/jpg":
		fallthrough
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}

	fileName := now + "_" + encodedNo + ext
	file, err := os.Create("files/" + fileName)
	defer file.Close()
	if err != nil {
		return
	}

	_, err = io.Copy(file, r.Body)
	if err != nil {
		return
	}

	w.WriteJson(map[string]string{
		"path": "/files/" + fileName,
	})
}
