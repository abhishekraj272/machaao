package machaao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

//MachaaoAPIToken Get MachaaoAPIToken from https://portal.messengerx.io
var MachaaoAPIToken string = ""

//WitAPIToken Get WitAPIToken from https://wit.ai
var WitAPIToken string = ""

//MachaaoBaseURL for dev, use https://ganglia-dev.machaao.com
var MachaaoBaseURL string = ""

//Server Starts server at given PORT. WebHook is machaao_hook
func Server() {
	port := GetPort()

	if WitAPIToken == "" {
		log.Fatalln("Wit API Token not initialised.")
	}
	if MachaaoAPIToken == "" {
		log.Fatalln("Machaao API Token not initialised.")
	}

	//API handler function
	http.HandleFunc("/machaao_hook", messageHandler)

	//Go http server
	log.Println("[-] Listening on...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

//SendPostReq Send post request
func SendPostReq(url string, body interface{}) (response *http.Response, err error) {

	//Body converted to json bytes from interface.
	jsonBody, _ := json.Marshal(body)

	//Post request sent to MessengerX.io API
	req, err1 := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))

	//Sets required headers for MessengerX.io API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_token", MachaaoAPIToken)

	if err1 != nil {
		panic(err1)
	}

	client := &http.Client{}
	resp, err2 := client.Do(req)

	if err2 != nil {
		panic(err2)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	bodyf, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(bodyf))

	return resp, nil
}

//GetPort Set PORT as env var or leave it to use 4747
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4747"
		log.Println("[-] No PORT environment variable detected. Setting to ", port)
	}
	return ":" + port
}

//Webhook messege handler
func messageHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	//This function reads the request Body and saves to body as byte.
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Error reading body: %v", err)
		return
	}

	//converts bytes to string
	var bodyData string = string(body)

	//incoming JWT Token
	var tokenString string = bodyData[8:(len(bodyData) - 2)]

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(MachaaoAPIToken), nil
	})

	if token == nil {
		log.Panicln("Incoming JWT Token invalid.")
		return
	}

	if err != nil {
		fmt.Println(err)
	}

	//captures message_data object from the JWT body.
	messageData := claims["sub"].(map[string]interface{})["messaging"].([]interface{})[0].(map[string]interface{})["message_data"]
	messageText := messageData.(map[string]interface{})["text"].(string)

	log.Println(messageData)
	log.Println(messageText)

	log.Println(r.Header["User_id"])

	return

}
