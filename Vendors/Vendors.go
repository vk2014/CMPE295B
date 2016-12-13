// Copyright Cisco Systems, Inc.
// All Rights Reserved.
// Author Vishnu Konepalli @ vkonepal@cisco.com
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"io/ioutil"
	"log/syslog"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func checkErr1(err error) {
	if err != nil {
		logwritter1.Notice("Error connecting to database")
		panic(err)
	}
}

func Vendors(w http.ResponseWriter, r *http.Request) {
	logwritter1.Notice("Processing PO info from Client IP Address: 128.107.1.74")
	fmt.Println("Processing PO info from Client IP Address: 128.107.1.74")
	var httpMethod = r.Method
	//fmt.Println(httpMethod)
	//RemoteIP := strings.Split(r.RemoteAddr, ":")
	//IPAddress := RemoteIP[0]
	//fmt.Println(IPAddress)
	if strings.EqualFold(httpMethod, "POST") {
		//var u Customer
		if r.Body == nil {
			logwritter1.Err("Unable to read HTTP Method")
			http.Error(w, "Please send a request body", 400)
			return
		}
		//err := json.NewDecoder(r.Body).Decode(&u)
		//body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logwritter1.Err("Unable to read HTTP Request Body")
			http.Error(w, err.Error(), 400)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {

		}
		//fmt.Println(string(body))

		var jsonBody map[string]string

		if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
			logwritter1.Err("Unable to read username from database")
			fmt.Print(err2.Error())
		}

		carLicensePlat := jsonBody["carLicensePlat"]
		parkingId := jsonBody["parkingId"]

		//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

		db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
		checkErr1(err)

		stmt, err := db.Prepare("INSERT purchaseorder SET carLicensePlat=?,parkingId=?")
		checkErr1(err)

		res, err := stmt.Exec(carLicensePlat, parkingId)
		checkErr1(err)

		res.LastInsertId()
		logwritter1.Notice("PO details for Licenseplate " + carLicensePlat + " Inserted")

		db.Close()

		var Vendors = map[string]string{}

		Vendors["status"] = "SUCCESS"
		Vendors["message"] = "PO info received and saved !!!"
		Vendors["messageCode"] = "200"

		//fmt.Println(Vendors)

		jsonInfo, err := json.Marshal(Vendors)
		if err != nil {
			logwritter1.Err("Error in PO JSON marchslling")
			fmt.Println("Error in PO JSON marchslling")
		}
		//fmt.Println(jsonInfo)
		//fmt.Println(Vendors)

		jsonInfo, err1 := json.Marshal(Vendors)

		//fmt.Println(jsonInfo)

		if err1 != nil {
			logwritter1.Err("Error in PO JSON marchslling")
			fmt.Println("Error in PO JSON marchslling")
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonInfo)
		S := string(jsonInfo)
		fmt.Println(S)

		w.Write([]byte("Thank You"))

	}
}

var logwritter1, err = syslog.New(syslog.LOG_ERR, "Vendors")

func main() {

	fmt.Println("Vendors started on port 9443.....")

	//logwritter,err := syslog.New(syslog.LOG_ERR,"CMPE295B")
	defer logwritter1.Close()
	if err != nil {
		log.Fatal("Error in the System")
	} else {
		logwritter1.Notice("Starting Server")

		http.HandleFunc("/", Vendors)

		//error := http.ListenAndServeTLS(":9443", "/Users/VKONEPAL/IdeaProjects/vkr/server.crt", "/Users/VKONEPAL/IdeaProjects/vkr/server.key", nil)
		//error := http.ListenAndServeTLS(":8443", "/home/cloud-user/go/src/github.com/CMPE295B/server.crt", "/home/cloud-user/go/src/github.com/CMPE295B/server.key", nil)
		error := http.ListenAndServe(":9443", nil)
		logwritter1.Err("Unable to Start Vendors Server")
		fmt.Println("Server finished .....")
		if err != nil {
			logwritter1.Alert(error.Error())
		}
	}

}
