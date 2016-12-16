// Copyright Cisco Systems, Inc.
// All Rights Reserved.
// Author Vishnu Konepalli @ vkonepal@cisco.com
// This package provides the server functionality for rendering different services for each client request
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math"
	"net/http"
	//	"strings"
	"encoding/json"
	"github.com/kellydunn/golang-geo"
	"log"
	"log/syslog"
	"strconv"
	//"github.com/gavv/gojsondiff/Godeps/_workspace/src/github.com/onsi/ginkgo/ginkgo/convert"
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
	//"os"
	//"github.com/streadway/amqp"
)

var regexProfile = regexp.MustCompile(`profile`)
var regexPrivacy = regexp.MustCompile(`privacy`)
var regexSmartParking = regexp.MustCompile(`smartparking`)
var regexDeleteUser = regexp.MustCompile(`user`)
var regexGetUser = regexp.MustCompile(`user`)
var regexPark = regexp.MustCompile(`park`)
var regexSensor = regexp.MustCompile(`park`)

type GetUserstruct struct {
	billingContact       string
	address              string
	email                string
	zipCode              string
	carLicensePlat       string
	shareLicencePlate    string
	shareParkingDuration string
	shareServiceUsages   string
	occupyTimeStamp      string
	leaveTimeStamp       string
	duration             string
	usageServices        string
	parkingId            string
}

type GetUserList []GetUserstruct

type Vendor struct {
	Distance   float64
	VendorName string
	CouponCode int32
}

type VendorList []Vendor

type User struct {
	fname   string
	lname   string
	emailid string
}

type UserList []User

type Customer struct {
	password string
	username string
	email    string
}

type CustomerResponse struct {
	status   string
	message  string
	username string
}

type customerList []CustomerResponse

type UserProfiles struct {
	profile Profiles
}

type Profiles struct {
	billingContact string
	address        string
	email          string
	zipCode        string
	carLicensePlat string
}

type ProfileResponse struct {
	status      string
	message     string
	messageCode string
}

type profileList []ProfileResponse

type Privacytype struct {
	shareLicencePlate string
}

type UserPrivacy struct {
	privacy Privacytype
}

func checkErr(err error) {
	if err != nil {
		logwritter.Notice("Error connecting to database")
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)

	name := r.URL.Query().Get("name")
	longitude := r.URL.Query().Get("longitude")
	latitude := r.URL.Query().Get("latitude")
	couponcode := r.URL.Query().Get("couponcode")
	biztype := r.URL.Query().Get("biztype")

	fmt.Fprintf(w, "Success %s!", r.URL.Path[1:])
	insertVendor(name, latitude, longitude, couponcode, biztype)
	logwritter.Notice("Vendor " + name + " Inserted")
}

func handlerUser(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)

	fname := r.URL.Query().Get("fname")
	fmt.Println(fname)
	lname := r.URL.Query().Get("lname")
	fmt.Println(lname)
	emailid := r.URL.Query().Get("emailid")
	password := r.URL.Query().Get("password")

	fmt.Fprintf(w, "Success %s!", r.URL.Path[1:])
	insertUser(fname, lname, emailid, password)
	logwritter.Notice("User " + fname + " Inserted")
}

func handlerServices(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)

	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")
	userService := r.URL.Query().Get("type")
	userRadius := r.URL.Query().Get("radius")

	fmt.Fprintf(w, "%s", getServices(userLatitude, userLongitude, userService, userRadius))
	logwritter.Notice("User requested Services " + userService + " rendered")
}

func handlerDefaultServices(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)

	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")

	fmt.Fprintf(w, "%s", getDefaultServices(userLatitude, userLongitude))
	logwritter.Notice("Default Services rendered")
}

// getServices API returns the services within distance range.
// API needs input paramaters like user Latitude, Longitude, requested Service and Radius
// API returns a JSON object with all the services found within the radius and service requested.
func getServices(uLatitude string, uLongitude string, uService, uRadius string) string {
	logwritter.Notice("Requesting Services")
	ult, err2 := strconv.ParseFloat(uLatitude, 64)
	if err2 != nil {
		logwritter.Err("Error in parsing getServices uLatitude")
		fmt.Println("Error in parsing getServices uLatitude ")
	}

	ulg, err1 := strconv.ParseFloat(uLongitude, 64)
	if err1 != nil {
		logwritter.Err("Error in parsing getServices uLongitude")
		fmt.Println("Error in parsing getServices uLongitude")
	}

	usrRadius, err2 := strconv.ParseFloat(uRadius, 64)
	if err2 != nil {
		logwritter.Err("Error in parsing getServices uRadius")
		fmt.Println("Error in parsing getServices uRadius")
	}

	//fmt.Println(ulg,ult,usrRadius)

	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	var result string = ""
	rows, err := db.Query("SELECT name,latitude,longitude,couponcode,biztype FROM vendors WHERE biztype=?", uService)
	logwritter.Notice("Reading user selected services from database")
	if err != nil {
		logwritter.Err("Error connecting to database")
		log.Fatal(err)

	}
	defer rows.Close()
	//var VendorId Vendor
	var Vendors = make(VendorList, 0)
	//var counter int = 0

	for rows.Next() {
		var name string
		var longitude float64
		var latitude float64
		var couponcode int32
		var biztype string
		if err := rows.Scan(&name, &latitude, &longitude, &couponcode, &biztype); err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", name)
		difference := Distance(latitude, longitude, ult, ulg)
		fmt.Println(difference, usrRadius, name)

		if difference < usrRadius {
			var vendor Vendor
			vendor.Distance = difference
			vendor.CouponCode = couponcode
			vendor.VendorName = name
			Vendors = append(Vendors, vendor)
			result = result + fmt.Sprintf("%f", difference) + " " + fmt.Sprintf("%s", name) + "; "
		}
	}
	if err := rows.Err(); err != nil {
		logwritter.Err("Error in fetching data from database")
		log.Fatal(err)
	}
	db.Close()

	jsonInfo, err := json.Marshal(Vendors)
	if err != nil {
		logwritter.Err("Error in getServices JSON marchslling")
		fmt.Println("Error in getServices JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	return S

}

func getDefaultServices(uLatitude string, uLongitude string) string {
	logwritter.Notice("Requesting Default Services")
	ult, err2 := strconv.ParseFloat(uLatitude, 64)
	if err2 != nil {
		logwritter.Err("Error in parsing getDefaultServices uLatitude")
		fmt.Println("Error in parsing getDefaultServices uLatitude ")
	}

	ulg, err1 := strconv.ParseFloat(uLongitude, 64)
	if err1 != nil {
		logwritter.Err("Error in parsing getDefaultServices uLongitude")
		fmt.Println("Error in parsing getDefaultServices uLongitude ")
	}

	//fmt.Println(ulg,ult,usrRadius)

	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	var result string = ""
	rows, err := db.Query("SELECT * from vendors")
	if err != nil {
		logwritter.Err("Error connecting to database")
		log.Fatal(err)
	}
	defer rows.Close()

	//var VendorId Vendor
	var Vendors = make(VendorList, 0)
	//var counter int = 0

	for rows.Next() {
		var name string
		var longitude float64
		var latitude float64
		var couponcode int32
		var biztype string
		var usrDefaultRadius float64 = 150.00
		if err := rows.Scan(&name, &latitude, &longitude, &couponcode, &biztype); err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", name)
		difference := Distance(latitude, longitude, ult, ulg)
		fmt.Println(difference, usrDefaultRadius, name)

		if difference < usrDefaultRadius {
			var vendor Vendor
			vendor.Distance = difference
			vendor.CouponCode = couponcode
			vendor.VendorName = name
			Vendors = append(Vendors, vendor)
			result = result + fmt.Sprintf("%f", difference) + " " + fmt.Sprintf("%s", name) + "; "
		}
	}
	if err := rows.Err(); err != nil {
		logwritter.Err("Error in fetching data from database")
		log.Fatal(err)
	}
	db.Close()
	fmt.Println(Vendors)

	jsonInfo, err := json.Marshal(Vendors)
	if err != nil {
		logwritter.Err("Error in getServices JSON marchslling")
		fmt.Println("Error in getDefaultServices JSON marchslling")
	}

	fmt.Println(jsonInfo)

	S := string(jsonInfo)
	fmt.Println(S)
	return S

	//return result

}

func hsin(theta float64) float64 {
	logwritter.Notice("Performing hsin calculation")
	return math.Pow(math.Sin(theta/2), 2)
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	logwritter.Notice("Performing distance calculation")
	// convert to radians
	// must cast radius as float to multiply later
	fmt.Println(lat1, lon1, lat2, lon2)

	/*var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))*/

	point1 := geo.NewPoint(lat1, lon1)
	point2 := geo.NewPoint(lat2, lon2)

	// find the great circle distance between them
	dist := point1.GreatCircleDistance(point2)
	return dist

}

func insertVendor(name string, latitude string, longitude string, couponcode string, biztype string) {
	logwritter.Notice("Requesting insert Vendor")
	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	stmt, err := db.Prepare("INSERT vendors SET name=?,latitude=?,longitude=?,couponcode=?,biztype=?")
	checkErr(err)

	res, err := stmt.Exec(name, latitude, longitude, couponcode, biztype)
	checkErr(err)

	res.LastInsertId()

	db.Close()
}

func insertUser(fname string, lname string, emailid string, password string) {
	logwritter.Notice("Requesting insert User")

	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	stmt, err := db.Prepare("INSERT users SET fname=?,lname=?,emailid=?,password=?")
	checkErr(err)

	res, err := stmt.Exec(fname, lname, emailid, password)
	checkErr(err)

	res.LastInsertId()

	var Users = make(UserList, 1)

	Users[0].fname = fname
	Users[0].lname = lname
	Users[0].emailid = emailid

	db.Close()

	jsonInfo, err := json.Marshal(Users)
	if err != nil {
		logwritter.Err("Error in insertUser JSON marchslling")
		fmt.Println("Error in JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	//return S

}

func AddUser(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting Add User")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//fmt.Println("Processing request from Client IP Address: "+IPAddress)
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logwritter.Err("Unable to read HTTP Request Body")
			http.Error(w, err.Error(), 400)
			return
		}

		var jsonBody map[string]string

		if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
			logwritter.Err("Error in AddUser JSON unmarchslling")
			fmt.Print(err2.Error())
		}

		username := jsonBody["username"]
		email := jsonBody["email"]
		password := jsonBody["password"]

		//fmt.Println(username,email,password)

		db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
		checkErr(err)

		stmt, err := db.Prepare("INSERT customers SET username=?,email=?,password=?")
		checkErr(err)

		res, err := stmt.Exec(username, email, password)
		checkErr(err)

		res.LastInsertId()
		logwritter.Notice("User with username " + username + " Inserted")
		logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Add user with username " + username)
		db.Close()

		var Vendors = map[string]string{}

		Vendors["status"] = "Added"
		Vendors["message"] = "User"
		Vendors["username"] = username

		//fmt.Println(Vendors)

		jsonInfo, err := json.Marshal(Vendors)
		if err != nil {
			logwritter.Err("Error in AddUser JSON marchslling")
			fmt.Println("Error in AddUser JSON marchslling")
		}
		//fmt.Println(jsonInfo)
		//fmt.Println(Vendors)

		jsonInfo, err1 := json.Marshal(Vendors)

		//fmt.Println(jsonInfo)

		if err1 != nil {
			fmt.Println("Error in AddUser JSON marchslling")
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonInfo)
		S := string(jsonInfo)
		//fmt.Printf("%+v", S)
		fmt.Println(S)
		//fmt.Fprintf(w, "%s",S)

	} else {
		logwritter.Err("AddUser - HTTP Method not supported")
		http.Error(w, "Method not supported", 400)
	}

}

func UserRoute(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Performing User Requests")

	fmt.Println(r.URL.Path)
	switch {
	case regexProfile.MatchString(r.URL.Path):
		Profile(w, r)
	case regexPrivacy.MatchString(r.URL.Path):
		Privacy(w, r)
	case regexSmartParking.MatchString(r.URL.Path):
		SmartParking(w, r)
	case regexDeleteUser.MatchString(r.URL.Path) && strings.EqualFold(r.Method, "DELETE"):
		DeleteUser(w, r)
	case regexGetUser.MatchString(r.URL.Path) && strings.EqualFold(r.Method, "GET"):
		GetUser(w, r)
	case regexSensor.MatchString(r.URL.Path) && strings.EqualFold(r.Method, "PUT"):
		Sensor(w, r)
	case regexPark.MatchString(r.URL.Path):
		Parking(w, r)
	default:
		w.Write([]byte("Unknown URL"))
	}

}

func Profile(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting Profile Add")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "user") && strings.EqualFold(arrTemp[3], "profile") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?", username)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {
				if err := rows.Scan(&name); err != nil {
					logwritter.Err("Unable to read username from database")
					log.Fatal(err)
				}

			}
			if err := rows.Err(); err != nil {
				logwritter.Err("Unable to read username from database")
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(username, name) {

				if r.Body == nil {
					http.Error(w, "Please send a request body", 400)
					return
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				var jsonBody map[string]string

				if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
					logwritter.Err("Error in Profile JSON unmarchslling")
					fmt.Print(err2.Error())
				}
				billingContact := jsonBody["billingContact"]
				address := jsonBody["address"]
				email := jsonBody["email"]
				zipcode := jsonBody["zipCode"]
				carLicensePlat := jsonBody["carLicensePlat"]

				//fmt.Println(billingContact, address, email, zipcode, carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT profiles SET billingContact=?,address=?,email=?,zipcode=?,carLicensePlat=?,username=?")
				checkErr(err)

				res, err := stmt.Exec(billingContact, address, email, zipcode, carLicensePlat, username)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Profile Add for username " + username)
				logwritter.Notice("Profile for username " + username + " Inserted")

				db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "Profile Inserted"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err := json.Marshal(Vendors)
				if err != nil {
					logwritter.Err("Error in Profile JSON marchslling")
					fmt.Println("Error in Profile JSON marchslling")
				}
				//fmt.Println(jsonInfo)
				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in Profile JSON marchslling")
					fmt.Println("Error in Profile JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)

			}

		}

	}
	// w.Write([]byte("DONE"))

}

func Privacy(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting Privacy Add")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "user") && strings.EqualFold(arrTemp[3], "privacy") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?", username)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {

				if err := rows.Scan(&name); err != nil {
					logwritter.Err("Unable to read username from database")
					log.Fatal(err)
				}

			}
			if err := rows.Err(); err != nil {
				logwritter.Err("Unable to read username from database")
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(username, name) {

				if r.Body == nil {
					http.Error(w, "Please send a request body", 400)
					return
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				var jsonBody map[string]string

				if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
					logwritter.Err("Unable to read username from database")
					fmt.Print(err2.Error())
				}

				shareLicencePlate := jsonBody["shareLicencePlate"]
				shareParkingDuration := jsonBody["shareParkingDuration"]
				shareServiceUsages := jsonBody["shareServiceUsages"]

				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT privacy SET shareLicencePlate=?,shareParkingDuration=?,username=?,shareServiceUsages=?")
				checkErr(err)

				res, err := stmt.Exec(shareLicencePlate, shareParkingDuration, username, shareServiceUsages)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Add Privacy details for username " + username)
				logwritter.Notice("Privacy details for username " + username + " Inserted")

				db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "Privacy Info Saved"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err := json.Marshal(Vendors)
				if err != nil {
					logwritter.Err("Error in Privacy JSON marchslling")
					fmt.Println("Error in Privacy JSON marchslling")
				}
				//fmt.Println(jsonInfo)
				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in Privacy JSON marchslling")
					fmt.Println("Error in Privacy JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)

			}

		}

	}
	//w.Write([]byte("DONE"))

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting Delete User")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "DELETE") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "user") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?", username)
			if err != nil {
				logwritter.Err("Unable to read username from database")
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {

				if err := rows.Scan(&name); err != nil {
					logwritter.Err("Unable to read username from database")
					log.Fatal(err)
				}

			}
			if err := rows.Err(); err != nil {
				logwritter.Err("Unable to read username from database")
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(username, name) {

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				//var result string = ""
				stmt1, err1 := db.Prepare("DELETE from customers where username=?")
				checkErr(err1)
				res1, err1 := stmt1.Exec(username)
				checkErr(err1)
				res1.RowsAffected()

				stmt2, err2 := db.Prepare("DELETE from profiles where username=?")
				checkErr(err2)
				res2, err2 := stmt2.Exec(username)
				checkErr(err2)
				res2.RowsAffected()

				stmt3, err3 := db.Prepare("DELETE from privacy where username=?")
				checkErr(err3)
				res3, err3 := stmt3.Exec(username)
				checkErr(err3)
				res3.RowsAffected()

				stmt4, err4 := db.Prepare("DELETE from smartparking where username=?")
				checkErr(err4)
				res4, err4 := stmt4.Exec(username)
				checkErr(err4)
				res4.RowsAffected()

				stmt5, err5 := db.Prepare("DELETE from parking where username=?")
				checkErr(err5)
				res5, err5 := stmt5.Exec(username)
				checkErr(err5)
				res5.RowsAffected()

				logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Delete request for username " + username)
				logwritter.Notice("User username " + username + " Deleted")

				db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "User Deleted"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in DeleteUser JSON marchslling")
					fmt.Println("Error in DeleteUser JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)

			}
		}

	}
	//w.Write([]byte("DONE"))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Performing Get User")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "GET") {
		arrTemp := strings.Split(r.URL.Path, "/")
		RemoteIP := strings.Split(r.RemoteAddr, ":")
		IPAddress := RemoteIP[0]
		fmt.Println("Processing request from Client IP Address: "+IPAddress)
		if strings.EqualFold(arrTemp[1], "user") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)
			logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Get details for username " + username)
			//var VendorId Vendor
			var Vendors = map[string]string{}

			//var counter int = 0
			//var vendor GetUserstruct

			rows1, err1 := db.Query("SELECT email FROM customers WHERE username=?", username)
			logwritter.Notice("Reading user data from database")
			if err1 != nil {
				logwritter.Err("Error connecting to database")
				log.Fatal(err1)

			}
			defer rows1.Close()
			for rows1.Next() {
				var email string

				if err1 := rows1.Scan(&email); err1 != nil {
					log.Fatal(err1)
				}
				Vendors["email"] = email
				//fmt.Println(Vendors)
			}
			if err1 := rows1.Err(); err1 != nil {
				logwritter.Err("Error in fetching data from database")
				log.Fatal(err1)
			}

			rows2, err2 := db.Query("SELECT billingContact,address,email,zipcode,carLicensePlat FROM profiles WHERE username=?", username)
			logwritter.Notice("Reading user data from database")
			if err2 != nil {
				logwritter.Err("Error connecting to database")
				log.Fatal(err2)

			}
			defer rows2.Close()
			for rows2.Next() {
				var billingContact string
				var address string
				var zipcode string
				var carLicensePlat string
				var email string

				if err2 := rows2.Scan(&billingContact, &address, &email, &zipcode, &carLicensePlat); err2 != nil {
					log.Fatal(err2)
				}
				Vendors["billingContact"] = billingContact
				Vendors["address"] = address
				Vendors["zipcode"] = zipcode
				Vendors["carLicensePlat"] = carLicensePlat
				//fmt.Println(Vendors)
			}
			if err2 := rows2.Err(); err2 != nil {
				logwritter.Err("Error in fetching data from database")
				log.Fatal(err2)
			}

			rows3, err3 := db.Query("SELECT shareLicencePlate,shareParkingDuration,shareServiceUsages FROM privacy WHERE username=?", username)
			logwritter.Notice("Reading user data from database")
			if err3 != nil {
				logwritter.Err("Error connecting to database")
				log.Fatal(err3)
			}
			defer rows3.Close()
			for rows3.Next() {
				var shareLicencePlate string
				var shareParkingDuration string
				var shareServiceUsages string

				if err3 := rows3.Scan(&shareLicencePlate, &shareParkingDuration, &shareServiceUsages); err3 != nil {
					log.Fatal(err3)
				}
				Vendors["shareLicencePlate"] = shareLicencePlate
				Vendors["shareParkingDuration"] = shareParkingDuration
				Vendors["shareServiceUsages"] = shareServiceUsages
				//fmt.Println(Vendors)
			}
			if err3 := rows3.Err(); err3 != nil {
				logwritter.Err("Error in fetching data from database")
				log.Fatal(err3)
			}

			rows4, err4 := db.Query("SELECT occupyTimeStamp,leaveTimeStamp,duration,parkingId,usageServices FROM smartparking WHERE username=?", username)
			logwritter.Notice("Reading user data from database")
			if err4 != nil {
				logwritter.Err("Error connecting to database")
				log.Fatal(err4)
			}
			defer rows4.Close()
			for rows4.Next() {
				var occupyTimeStamp string
				var leaveTimeStamp string
				var duration string
				var parkingId string
				var usageServices string

				if err4 := rows4.Scan(&occupyTimeStamp, &leaveTimeStamp, &duration, &parkingId, &usageServices); err4 != nil {
					log.Fatal(err4)
				}

				Vendors["occupyTimeStamp"] = occupyTimeStamp
				Vendors["leaveTimeStamp"] = leaveTimeStamp
				Vendors["duration"] = duration
				Vendors["parkingId"] = parkingId
				Vendors["usageServices"] = usageServices
				//fmt.Println(Vendors)
			}
			if err4 := rows4.Err(); err4 != nil {
				logwritter.Err("Error in fetching data from database")
				log.Fatal(err4)
			}

			db.Close()

			jsonInfo, err := json.Marshal(Vendors)
			if err != nil {
				logwritter.Err("Error in GetUser JSON marchslling")
				fmt.Println("Error in GetUser JSON marchslling")
			}

			S := string(jsonInfo)
			fmt.Println(S)
			w.Write(jsonInfo)
			//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Get details for username " + username)
			logwritter.Notice("Get User details rendered")

			//return S
		}

	}
	//w.Write([]byte("DONE"))

}

func SmartParking(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting SmartParking")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "user") && strings.EqualFold(arrTemp[3], "smartparking") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?", username)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {

				if err := rows.Scan(&name); err != nil {
					log.Fatal(err)
				}
			}
			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(username, name) {

				if r.Body == nil {
					http.Error(w, "Please send a request body", 400)
					return
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				var jsonBody map[string]string

				if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
					fmt.Print(err2.Error())
				}

				occupyTimeStamp := jsonBody["occupyTimeStamp"]
				leaveTimeStamp := jsonBody["leaveTimeStamp"]
				duration := jsonBody["duration"]
				parkingId := jsonBody["parkingId"]
				usageServices := jsonBody["usageServices"]

				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT smartparking SET occupyTimeStamp=?,leaveTimeStamp=?,duration=?,parkingId=?,username=?,usageServices=?")
				checkErr(err)

				res, err := stmt.Exec(occupyTimeStamp, leaveTimeStamp, duration, parkingId, username, usageServices)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Parking information for username " + username)
				logwritter.Notice("Parking info for username " + username + " Inserted")

				//db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "Parking Info Saved"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err := json.Marshal(Vendors)
				if err != nil {
					logwritter.Err("Error in SmartParking JSON marchslling")
					fmt.Println("Error in SmartParking JSON marchslling")
				}
				//fmt.Println(jsonInfo)
				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in SmartParking JSON marchslling")
					fmt.Println("Error in SmartParking JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)

				rows1, err1 := db.Query("SELECT carLicensePlat FROM profiles WHERE username=?", username)
				logwritter.Notice("Reading user data from database")
				if err1 != nil {
					logwritter.Err("Error connecting to database")
					log.Fatal(err1)

				}

				defer rows1.Close()
				var carLicensePlat string
				for rows1.Next() {

					if err1 := rows1.Scan(&carLicensePlat); err1 != nil {
						log.Fatal(err1)
					}

					//fmt.Println(Vendors)
				}
				if err1 := rows1.Err(); err1 != nil {
					logwritter.Err("Error in fetching data from database")
					log.Fatal(err1)
				}
				if strings.EqualFold(usageServices, "") == false {
					logwritter.Notice("Sending PO as usageServices has be chosen!")
					SendPO("http://127.0.0.1:9443/", carLicensePlat, parkingId)
				}

				db.Close()

			}

		}

	}
	//w.Write([]byte("DONE"))
}

func SendPO(url string, carLicensePlat string, parkingId string) {

	var PO = map[string]string{}

	PO["carLicensePlat"] = carLicensePlat
	PO["parkingId"] = parkingId

	//fmt.Println(Vendors)

	jsonInfo, err := json.Marshal(PO)
	if err != nil {
		logwritter.Err("Error in SendPO JSON marchslling")
		fmt.Println("Error in SendPO JSON marchslling")
	}

	//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonInfo))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func Parking(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Requesting Parking")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "POST") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "user") && strings.EqualFold(arrTemp[3], "park") {
			username := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?", username)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {

				if err := rows.Scan(&name); err != nil {
					log.Fatal(err)
				}
			}
			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(username, name) {

				if r.Body == nil {
					http.Error(w, "Please send a request body", 400)
					return
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				var jsonBody map[string]string

				if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
					fmt.Print(err2.Error())
				}

				Parkingid := jsonBody["Parkingid"]

				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT parking SET Parkingid=?,username=?")
				checkErr(err)

				res, err := stmt.Exec(Parkingid, username)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Processing request from  Client IP Address: "+IPAddress + " Parking Parkingid: "+Parkingid +" chosen by user " + username)
				logwritter.Notice("Parking info associated with username " + username + " Inserted")

				db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "Parking Info Saved"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err := json.Marshal(Vendors)
				if err != nil {
					logwritter.Err("Error in Parking JSON marchslling")
					fmt.Println("Error in Parking JSON marchslling")
				}
				//fmt.Println(jsonInfo)
				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in Parking JSON marchslling")
					fmt.Println("Error in Parking JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)
			}

		}

	}
	//w.Write([]byte("DONE"))
}

func Sensor(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Processing Sensor Data")
	RemoteIP := strings.Split(r.RemoteAddr, ":")
	IPAddress := RemoteIP[0]
	//logwritter.Notice("Processing request from  Client IP Address: "+IPAddress)
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path, "/")
		if strings.EqualFold(arrTemp[1], "park") {
			Parkingid := arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT Parkingid FROM parking WHERE Parkingid=?", Parkingid)
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()

			var name string
			for rows.Next() {

				if err := rows.Scan(&name); err != nil {
					log.Fatal(err)
				}
			}
			if err := rows.Err(); err != nil {
				log.Fatal(err)
			}
			db.Close()

			if strings.EqualFold(Parkingid, name) {

				if r.Body == nil {
					http.Error(w, "Please send a request body", 400)
					return
				}
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), 400)
					return
				}

				var jsonBody map[string]string

				if err2 := json.Unmarshal(body, &jsonBody); err2 != nil {
					fmt.Print(err2.Error())
				}

				Occupied := jsonBody["Occupied"]

				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("UPDATE parking SET Parkingid=?,Occupied=?")
				checkErr(err)

				res, err := stmt.Exec(Parkingid, Occupied)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Sensor Data for Parkingid " + Parkingid + " updated" + " from Client IP Address: " + IPAddress)
				//logwritter.Notice("Sensor Data for Parkingid " + Parkingid + " updated")

				db.Close()

				var Vendors = map[string]string{}

				Vendors["status"] = "OK"
				Vendors["message"] = "Sensor Info Saved"
				Vendors["messageCode"] = "200"

				//fmt.Println(Vendors)

				jsonInfo, err := json.Marshal(Vendors)
				if err != nil {
					logwritter.Err("Error in Sensor JSON marchslling")
					fmt.Println("Error in Sensor JSON marchslling")
				}
				//fmt.Println(jsonInfo)
				//fmt.Println(Vendors)

				jsonInfo, err1 := json.Marshal(Vendors)

				//fmt.Println(jsonInfo)

				if err1 != nil {
					logwritter.Err("Error in Sensor JSON marchslling")
					fmt.Println("Error in Sensor JSON marchslling")
				}

				w.Header().Add("Content-Type", "application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)
			}

		}

	}
	//w.Write([]byte("DONE"))
}

var logwritter, err = syslog.New(syslog.LOG_ERR, "CMPE295B")

func main() {

	fmt.Println("Server started on port 8443.....")

	//logwritter,err := syslog.New(syslog.LOG_ERR,"CMPE295B")
	defer logwritter.Close()
	if err != nil {
		log.Fatal("Error in the System")
	} else {
		logwritter.Notice("Starting Server")
		http.HandleFunc("/insertvendor", handler)
		http.HandleFunc("/insertuser", handlerUser)
		http.HandleFunc("/getservices", handlerServices)
		http.HandleFunc("/getdefaultservices", handlerDefaultServices)
		http.HandleFunc("/user", AddUser)
		//http.HandleFunc("/addprofile", AddProfile)
		http.HandleFunc("/", UserRoute)

		//error := http.ListenAndServeTLS(":8443", "/Users/VKONEPAL/IdeaProjects/vkr/server.crt", "/Users/VKONEPAL/IdeaProjects/vkr/server.key", nil)
		error := http.ListenAndServeTLS(":8443", "/home/cloud-user/go/src/github.com/CMPE295B/server.crt", "/home/cloud-user/go/src/github.com/CMPE295B/server.key", nil)
		logwritter.Err("Unable to Start Server")
		fmt.Println("Server finished 456.....")
		if err != nil {
			logwritter.Alert(error.Error())
		}
	}

}