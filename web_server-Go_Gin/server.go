package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	UrlOfThisServer := "http://localhost:8080/"
	UrlOfAPI := "http://localhost:8000/"

	beginCode := "x&z@2020$10$17#messagebegin:"
	endCode := "x&z@2020$10$17#messageend:"

	binaryBeginCode := eightBitBinaryEncoder(beginCode)
	binaryEndCode := eightBitBinaryEncoder(endCode)

	// fmt.Println(binaryBeginCode)

	var filename string
	var binaryExistedMessage string
	var UrlOfImage string

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})

	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			// fmt.Println(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		filename = filepath.Base(file.Filename)
		currentFilepath := "./static/" + filename
		if err := c.SaveUploadedFile(file, currentFilepath); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			// fmt.Println(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		UrlOfImage = UrlOfThisServer + "static/" + filename

		capacity := getMessageCapacity(UrlOfImage, UrlOfAPI+"messagecapacity", binaryBeginCode, binaryEndCode)
		capacity /= 8

		var isMessageExist string

		if UrlOfImage[len(UrlOfImage)-3:] != "png" {
			isMessageExist = "false"
		} else {
			if checkMessage(UrlOfImage, UrlOfAPI+"checkmessage", binaryBeginCode) {
				isMessageExist = "true"
			} else {
				isMessageExist = "false"
			}
		}

		data := map[string]interface{}{
			"filename":       filename,
			"capacity":       capacity,
			"isMessageExist": isMessageExist,
		}

		c.JSONP(http.StatusOK, data)
	})

	router.POST("/hidemessage", func(c *gin.Context) {
		type JsonObj struct {
			Message string
			Method  string
		}

		var messageJson JsonObj

		if err := c.ShouldBindJSON(&messageJson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var imageUrl string
		binaryMessage := eightBitBinaryEncoder(messageJson.Message)

		if messageJson.Method == "replace" {
			imageUrl = hideMessage(UrlOfImage, UrlOfAPI+"hidemessage", binaryBeginCode, binaryMessage, binaryEndCode)
		} else {
			binaryMessage = eightBitBinaryEncoder("\n") + binaryMessage
			imageUrl = appendMessage(UrlOfImage, UrlOfAPI+"appendmessage", binaryBeginCode, binaryExistedMessage, binaryMessage, binaryEndCode)
		}

		imageUrl = UrlOfAPI + imageUrl

		data := map[string]interface{}{
			"imageUrl": imageUrl,
		}
		c.JSONP(http.StatusOK, data)
	})

	router.POST("/erasemessage", func(c *gin.Context) {
		var imageUrl string
		imageUrl = eraseMessage(UrlOfImage, UrlOfAPI+"erasemessage", binaryBeginCode, binaryExistedMessage, binaryEndCode)
		imageUrl = UrlOfAPI + imageUrl

		data := map[string]interface{}{
			"imageUrl": imageUrl,
		}
		c.JSONP(http.StatusOK, data)
	})

	router.POST("/readmessage", func(c *gin.Context) {
		binaryExistedMessage = readMessage(UrlOfImage, UrlOfAPI+"readmessage", binaryBeginCode, binaryEndCode)
		message := eightBitBinaryDecoder(binaryExistedMessage)

		data := map[string]interface{}{
			"message": message,
		}

		c.JSONP(http.StatusOK, data)
	})

	router.Static("/static", "./static")
	router.Run(":8080")
}

func eightBitBinaryEncoder(message string) string {
	binaryCode := ""
	for i := 0; i < len(message); i++ {
		asciiValue := int(message[i])

		powOf2 := 128
		for i := 0; i < 8; i++ {
			if asciiValue >= powOf2 {
				binaryCode += "1"
				asciiValue -= powOf2
			} else {
				binaryCode += "0"
			}
			powOf2 /= 2
		}
	}
	return binaryCode
}

func eightBitBinaryDecoder(binaryCode string) string {
	message := ""
	for i := 0; i < len(binaryCode); i += 8 {
		block := binaryCode[i : i+8]
		asciiValue := 0

		powOf2 := 128
		for i := 0; i < 8; i++ {
			if block[i] == '1' {
				asciiValue += powOf2
			}
			powOf2 /= 2
		}

		character := string(rune(asciiValue))
		message += character
	}
	return message
}

func getMessageCapacity(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string, binaryEndCode string) int {
	requestBody, err := json.Marshal(map[string]string{
		"image_url":       UrlOfImage,
		"binaryBeginCode": binaryBeginCode,
		"binaryEndCode":   binaryEndCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		MessageCapacity int
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.MessageCapacity
}

func checkMessage(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string) bool {

	requestBody, err := json.Marshal(map[string]string{
		"image_url":       UrlOfImage,
		"binaryBeginCode": binaryBeginCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		IsMessageExist bool
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.IsMessageExist
}

func hideMessage(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string, binaryMessage string, binaryEndCode string) string {
	requestBody, err := json.Marshal(map[string]string{
		"image_url":       UrlOfImage,
		"binaryBeginCode": binaryBeginCode,
		"binaryMessage":   binaryMessage,
		"binaryEndCode":   binaryEndCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		NewImageURL string
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.NewImageURL
}

func appendMessage(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string, binaryExistedMessage string, binaryMessage string, binaryEndCode string) string {
	requestBody, err := json.Marshal(map[string]string{
		"image_url":            UrlOfImage,
		"binaryBeginCode":      binaryBeginCode,
		"binaryExistedMessage": binaryExistedMessage,
		"binaryMessage":        binaryMessage,
		"binaryEndCode":        binaryEndCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		NewImageURL string
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.NewImageURL
}

func eraseMessage(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string, binaryExistedMessage string, binaryEndCode string) string {
	requestBody, err := json.Marshal(map[string]string{
		"image_url":            UrlOfImage,
		"binaryBeginCode":      binaryBeginCode,
		"binaryExistedMessage": binaryExistedMessage,
		"binaryEndCode":        binaryEndCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		NewImageURL string
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.NewImageURL
}

func readMessage(UrlOfImage string, UrlToBeRequest string, binaryBeginCode string, binaryEndCode string) string {
	requestBody, err := json.Marshal(map[string]string{
		"image_url":       UrlOfImage,
		"binaryBeginCode": binaryBeginCode,
		"binaryEndCode":   binaryEndCode,
	})

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(UrlToBeRequest, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	type JsonObj struct {
		BinaryMessage string
	}

	jsonData := []byte(body)

	var jsonObj JsonObj
	err = json.Unmarshal(jsonData, &jsonObj)

	if err != nil {
		log.Println(err)
	}

	return jsonObj.BinaryMessage
}
