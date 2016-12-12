package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"log/syslog"
	"io/ioutil"
)

func Vendors(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Vendors")
	var httpMethod = r.Method
	//fmt.Println(httpMethod)

	if strings.EqualFold(httpMethod, "POST") {
		//var u Customer
		if r.Body == nil {
			logwritter.Err("Unable to read HTTP Method")
			http.Error(w, "Please send a request body", 400)
			return
		}
		//err := json.NewDecoder(r.Body).Decode(&u)
		//body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logwritter.Err("Unable to read HTTP Request Body")
			http.Error(w, err.Error(), 400)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {

		}
		fmt.Println(string(body))
		w.Write([]byte("Thank You"))

	}
}

var logwritter, err = syslog.New(syslog.LOG_ERR,"Vendors")

func main(){

	fmt.Println("Vendors started .....")

	//logwritter,err := syslog.New(syslog.LOG_ERR,"CMPE295B")
	defer logwritter.Close()
	if err != nil {
		log.Fatal("Error in the System")
	}else {
		logwritter.Notice("Starting Server")

		http.HandleFunc("/", Vendors)

		//error := http.ListenAndServeTLS(":9443", "/Users/VKONEPAL/IdeaProjects/vkr/server.crt", "/Users/VKONEPAL/IdeaProjects/vkr/server.key", nil)
		//error := http.ListenAndServeTLS(":8443", "/home/cloud-user/go/src/github.com/CMPE295B/server.crt", "/home/cloud-user/go/src/github.com/CMPE295B/server.key", nil)
		error := http.ListenAndServe(":9443",nil)
		logwritter.Err("Unable to Start Server")
		fmt.Println("Server finished .....")
		if err != nil {
			logwritter.Alert(error.Error())
		}
	}

}
