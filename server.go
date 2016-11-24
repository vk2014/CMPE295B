package main

import (
	"fmt"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"math"
//	"strings"
	"strconv"
	"github.com/kellydunn/golang-geo"
	"log"
	"encoding/json"
	"log/syslog"
	//"github.com/gavv/gojsondiff/Godeps/_workspace/src/github.com/onsi/ginkgo/ginkgo/convert"

	"io/ioutil"
	//"github.com/revel/modules/db/app"
	"strings"
	"regexp"
)

var regexProfile = regexp.MustCompile(`profile`)
var regexPrivacy = regexp.MustCompile(`privacy`)
var regexSmartParking = regexp.MustCompile(`smartparking`)
var regexDeleteUser = regexp.MustCompile(`user`)
var regexGetUser = regexp.MustCompile(`user`)

type Vendor struct {
	Distance float64
	VendorName string
	CouponCode int32
}

type VendorList []Vendor

type User struct {
	fname string
	lname string
	emailid string
}

type UserList []User

type Customer struct {
	password string
	username string
	email string
}

type CustomerResponse struct {
	status string
	message string
	username string
}

type customerList []CustomerResponse

type UserProfiles struct {
	profile Profiles
}

type Profiles struct {

	billingContact string
	address string
	email string
	zipCode string
	carLicensePlat string

}

type ProfileResponse struct {
status string
message string
messageCode string
}

type profileList []ProfileResponse

type Privacytype struct {
	shareLicencePlate string
}

type UserPrivacy struct {
	privacy Privacytype
}

/*type VendorList struct {
	Vendors []Vendor
}*/

func checkErr(err error) {
	if err != nil {
		logwritter.Notice("Error connecting to database")
		panic(err)
	}
}

type test_struct struct {
	Test string
}

func parsePost(rw http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	fmt.Println(decoder)
/*
	var t test_struct
	err := json.Unmarshal(decoder,&t)

	if err != nil {
		panic(err)
	}
	fmt.Println(t)
	fmt.Println(t.Test)
*/
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	name := r.URL.Query().Get("name")
	longitude := r.URL.Query().Get("longitude")
	latitude := r.URL.Query().Get("latitude")
	couponcode := r.URL.Query().Get("couponcode")
	biztype := r.URL.Query().Get("biztype")

	fmt.Fprintf(w, "Success %s!", r.URL.Path[1:])
	insertVendor(name,latitude,longitude,couponcode,biztype)
	logwritter.Notice("Vendor "+name+" Inserted")
}

func handlerUser(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	fname := r.URL.Query().Get("fname")
	fmt.Println(fname)
	lname := r.URL.Query().Get("lname")
	fmt.Println(lname)
	emailid := r.URL.Query().Get("emailid")
	password := r.URL.Query().Get("password")

	fmt.Fprintf(w, "Success %s!", r.URL.Path[1:])
	insertUser(fname,lname,emailid,password)
	logwritter.Notice("User "+fname+" Inserted")
}

func handlerServices(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")
	userService := r.URL.Query().Get("type")
	userRadius := r.URL.Query().Get("radius")

	fmt.Fprintf(w, "%s", getServices(userLatitude,userLongitude,userService,userRadius))
	logwritter.Notice("User requested Services "+userService+" rendered")
}

func handlerDefaultServices(w http.ResponseWriter, r *http.Request) {
	//fmt.Print(db.Ping())
	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")

	fmt.Fprintf(w, "%s", getDefaultServices(userLatitude,userLongitude))
	logwritter.Notice("Default Services rendered")
}

func getServices(uLatitude string,uLongitude string,uService,uRadius string) string {
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
	rows, err := db.Query("SELECT name,latitude,longitude,couponcode,biztype FROM vendors WHERE biztype=?",uService )
	logwritter.Notice("Reading user selected services from database")
	if err != nil {
		logwritter.Err("Error connecting to database")
		log.Fatal(err)

	}
	defer rows.Close()
	//var VendorId Vendor
	var Vendors = make(VendorList,0)
	//var counter int = 0

	for rows.Next() {
		var name string
		var longitude float64
		var latitude float64
		var couponcode int32
		var biztype string
		if err := rows.Scan(&name,&latitude,&longitude,&couponcode,&biztype); err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", name)
		difference := Distance(latitude,longitude,ult,ulg)
		fmt.Println(difference,usrRadius,name)

		if(difference<usrRadius) {
			var vendor Vendor
			vendor.Distance=difference
			vendor.CouponCode=couponcode
			vendor.VendorName=name
			Vendors = append(Vendors,vendor)
			result = result + fmt.Sprintf("%f", difference) + " " + fmt.Sprintf("%s", name) + "; "
		}
	}
	if err := rows.Err(); err != nil {
		logwritter.Err("Error in fetching data from database")
		log.Fatal(err)
	}
	db.Close()

	jsonInfo,err := json.Marshal(Vendors)
	if err != nil {
		logwritter.Err("Error in getServices JSON marchslling")
		fmt.Println("Error in getServices JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	return S

}

func getDefaultServices(uLatitude string,uLongitude string) string {
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
	rows, err := db.Query("SELECT * from vendors" )
	if err != nil {
		logwritter.Err("Error connecting to database")
		log.Fatal(err)
	}
	defer rows.Close()

	//var VendorId Vendor
	var Vendors = make(VendorList,0)
	//var counter int = 0

	for rows.Next() {
		var name string
		var longitude float64
		var latitude float64
		var couponcode int32
		var biztype string
		var usrDefaultRadius float64 = 150.00
		if err := rows.Scan(&name,&latitude,&longitude,&couponcode,&biztype); err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s\n", name)
		difference := Distance(latitude,longitude,ult,ulg)
		fmt.Println(difference,usrDefaultRadius,name)

		if(difference<usrDefaultRadius) {
			var vendor Vendor
			vendor.Distance=difference
			vendor.CouponCode=couponcode
			vendor.VendorName=name
			Vendors = append(Vendors,vendor)
			result = result + fmt.Sprintf("%f", difference) + " " + fmt.Sprintf("%s", name) + "; "
		}
	}
	if err := rows.Err(); err != nil {
		logwritter.Err("Error in fetching data from database")
		log.Fatal(err)
	}
	db.Close()
	fmt.Println(Vendors)

	jsonInfo,err := json.Marshal(Vendors)
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

/*
func AddProfile1(w http.ResponseWriter, r *http.Request) {

	var httpMethod = r.Method
	fmt.Println(httpMethod)

	if strings.EqualFold(httpMethod, "PUT") {
		//var u Customer
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}
		//err := json.NewDecoder(r.Body).Decode(&u)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		var str = r.RequestURI
		fmt.Println(str)
		fmt.Println(body)

	}
}
*/

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
	point1 := geo.NewPoint(lat1,lon1)
	point2 := geo.NewPoint(lat2,lon2)


	// find the great circle distance between them
	dist := point1.GreatCircleDistance(point2)
	return dist;

}

/*func getList(requestbiztype string,requestdistance float64,userLongitude float64,userLatitude float64) string {
	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)


	rows,err := db.Query("select * from vendors")
	checkErr(err)

	for rows.Next(){
		var name string
		var longitude float64
		var latitude float64
		var couponcode int
		var biztype string
		err = rows.Scan(&name,&longitude,&latitude,&couponcode,&biztype)
		checkErr(err)

		if strings.EqualFold(biztype,requestbiztype){
			var distance = Distance(longitude,latitude,userLongitude,userLatitude)
			if distance <= requestdistance{

			}
		}

	}



}*/

func insertVendor(name string,latitude string,longitude string,couponcode string,biztype string) {
	logwritter.Notice("Requesting insert Vendor")
	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	stmt, err := db.Prepare("INSERT vendors SET name=?,latitude=?,longitude=?,couponcode=?,biztype=?")
	checkErr(err)

	res, err := stmt.Exec(name,latitude,longitude,couponcode,biztype)
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

	res, err := stmt.Exec(fname,lname,emailid,password)
	checkErr(err)

	res.LastInsertId()

	var Users = make(UserList,1)


	Users[0].fname=fname
	Users[0].lname=lname
	Users[0].emailid=emailid

	db.Close()

	jsonInfo,err := json.Marshal(Users)
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
	var httpMethod = r.Method
	//fmt.Println(httpMethod)

	if strings.EqualFold(httpMethod,"POST") {
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
		logwritter.Notice("User with username "+username+" Inserted")
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

		w.Header().Add("Content-Type","application/json")
		w.Write(jsonInfo)
		S := string(jsonInfo)
		//fmt.Printf("%+v", S)
		fmt.Println(S)
		//fmt.Fprintf(w, "%s",S)

	}else{
		logwritter.Err("AddUser - HTTP Method not supported")
		http.Error(w,"Method not supported",400)
	}


}

func UserRoute(w http.ResponseWriter, r *http.Request) {
	logwritter.Notice("Performing User Requests")

	/*var httpMethod = r.Method
	fmt.Println(httpMethod)

	if strings.EqualFold(httpMethod, "PUT") {
	//var u Customer
	if r.Body == nil {
	http.Error(w, "Please send a request body", 400)
	return
	}
	//err := json.NewDecoder(r.Body).Decode(&u)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
	http.Error(w, err.Error(), 400)
	return
	}

	var str = r.RequestURI
	fmt.Println(str)
	fmt.Println(body)
	}
	*/

	fmt.Println(r.URL.Path)
	switch {
	case regexProfile.MatchString(r.URL.Path):
		Profile(w, r)
	case regexPrivacy.MatchString(r.URL.Path):
		Privacy(w, r)
	case regexSmartParking.MatchString(r.URL.Path):
		SmartParking(w, r)
	case regexDeleteUser.MatchString(r.URL.Path) && strings.EqualFold(r.Method,"DELETE"):
		DeleteUser(w, r)
	case regexGetUser.MatchString(r.URL.Path) && strings.EqualFold(r.Method,"GET"):
		GetUser(w, r)
	default:
		w.Write([]byte("Unknown URL"))
	}

}

func Profile(w http.ResponseWriter, r *http.Request){
	logwritter.Notice("Requesting Profile Add")
	if strings.EqualFold(r.Method, "PUT") {
	      arrTemp := strings.Split(r.URL.Path,"/")
		if strings.EqualFold(arrTemp[1],"user") && strings.EqualFold(arrTemp[3],"profile"){
			username :=arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?",username)
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

			if strings.EqualFold(username,name) {

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

				fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT profiles SET billingContact=?,address=?,email=?,zipcode=?,carLicensePlat=?,username=?")
				checkErr(err)

				res, err := stmt.Exec(billingContact, address, email, zipcode, carLicensePlat,username)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Profile for username "+username+" Inserted")

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

				w.Header().Add("Content-Type","application/json")
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

func Privacy(w http.ResponseWriter, r *http.Request){
	logwritter.Notice("Requesting Privacy Add")
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path,"/")
		if strings.EqualFold(arrTemp[1],"user") && strings.EqualFold(arrTemp[3],"privacy"){
			username :=arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?",username)
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

			if strings.EqualFold(username,name) {

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


				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)


				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT privacy SET shareLicencePlate=?,shareParkingDuration=?,username=?")
				checkErr(err)

				res, err := stmt.Exec(shareLicencePlate, shareParkingDuration,username)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Privacy details for username "+username+" Inserted")

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

				w.Header().Add("Content-Type","application/json")
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

func DeleteUser(w http.ResponseWriter, r *http.Request){
	logwritter.Notice("Requesting Delete User")
	if strings.EqualFold(r.Method, "DELETE") {
		arrTemp := strings.Split(r.URL.Path,"/")
		if strings.EqualFold(arrTemp[1],"user") {
			username :=arrTemp[2]
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

			logwritter.Notice("User username "+username+" Deleted")

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

				w.Header().Add("Content-Type","application/json")
				w.Write(jsonInfo)
				S := string(jsonInfo)
				//fmt.Printf("%+v", S)
				fmt.Println(S)
				//fmt.Fprintf(w, "%s",S)

		}

	}
	w.Write([]byte("DONE"))

}

func GetUser(w http.ResponseWriter, r *http.Request){
	logwritter.Notice("Requesting Get User")
	if strings.EqualFold(r.Method, "GET") {
		arrTemp := strings.Split(r.URL.Path,"/")
		if strings.EqualFold(arrTemp[1],"user") {
			username :=arrTemp[2]
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

			logwritter.Notice("User "+username+" detailed rendered")

			db.Close()

			var Vendors = map[string]string{}

			Vendors["status"] = "OK"
			Vendors["message"] = "User Deleted"
			Vendors["messageCode"] = "200"

			//fmt.Println(Vendors)

			jsonInfo, err1 := json.Marshal(Vendors)

			//fmt.Println(jsonInfo)

			if err1 != nil {
				logwritter.Err("Error in GetUser JSON marchslling")
				fmt.Println("Error in GetUser JSON marchslling")
			}

			w.Header().Add("Content-Type","application/json")
			w.Write(jsonInfo)
			S := string(jsonInfo)
			//fmt.Printf("%+v", S)
			fmt.Println(S)
			//fmt.Fprintf(w, "%s",S)



		}

	}
	w.Write([]byte("DONE"))

}

func SmartParking(w http.ResponseWriter, r *http.Request){
	logwritter.Notice("Requesting SmartParking")
	if strings.EqualFold(r.Method, "PUT") {
		arrTemp := strings.Split(r.URL.Path,"/")
		if strings.EqualFold(arrTemp[1],"user") && strings.EqualFold(arrTemp[3],"smartparking"){
			username :=arrTemp[2]
			db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
			checkErr(err)

			//var result string = ""
			rows, err := db.Query("SELECT username FROM customers WHERE username=?",username)
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

			if strings.EqualFold(username,name) {

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

				//fmt.Println(billingContact,address,email,zipcode,carLicensePlat)

				db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
				checkErr(err)

				stmt, err := db.Prepare("INSERT smartparking SET occupyTimeStamp=?,leaveTimeStamp=?,duration=?,parkingId=?,username=?")
				checkErr(err)

				res, err := stmt.Exec(occupyTimeStamp, leaveTimeStamp, duration, parkingId,username)
				checkErr(err)

				res.LastInsertId()
				logwritter.Notice("Parking info for username "+username+" Inserted")

				db.Close()

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
					fmt.Println("Error in getDefaultServices JSON marchslling")
				}

				w.Header().Add("Content-Type","application/json")
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

var logwritter, err = syslog.New(syslog.LOG_ERR,"CMPE295B")

func main() {

	fmt.Println("Server started 123.....")

	//logwritter,err := syslog.New(syslog.LOG_ERR,"CMPE295B")
	defer logwritter.Close()
	if err != nil {
		log.Fatal("Error in the System")
	}else {
		logwritter.Notice("Starting Server")
		http.HandleFunc("/insertvendor", handler)
		http.HandleFunc("/insertuser", handlerUser)
		http.HandleFunc("/getservices", handlerServices)
		http.HandleFunc("/getdefaultservices", handlerDefaultServices)
		http.HandleFunc("/user", AddUser)
		//http.HandleFunc("/addprofile", AddProfile)
		http.HandleFunc("/", UserRoute)

		error := http.ListenAndServeTLS(":8443", "/Users/VKONEPAL/IdeaProjects/vkr/server.crt", "/Users/VKONEPAL/IdeaProjects/vkr/server.key", nil)
		//error := http.ListenAndServeTLS(":8443", "/home/cloud-user/go/src/github.com/CMPE295B/server.crt", "/home/cloud-user/go/src/github.com/CMPE295B/server.key", nil)
		logwritter.Err("Unable to Start Server")
		fmt.Println("Server finished 456.....")
		if err != nil {
			logwritter.Alert(error.Error())
		}
	}

}