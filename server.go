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
)

type Vendor struct {
	Distance float64
	VendorName string
	CouponCode int32
}

type User struct {
	fname string
	lname string
	emailid string
}

/*type VendorList struct {
	Vendors []Vendor
}*/

type VendorList []Vendor

type UserList []User

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
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
}



func handlerServices(w http.ResponseWriter, r *http.Request) {


	//fmt.Print(db.Ping())
	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")
	userService := r.URL.Query().Get("type")
	userRadius := r.URL.Query().Get("radius")

	fmt.Fprintf(w, "%s", getServices(userLatitude,userLongitude,userService,userRadius))
}

func handlerDefaultServices(w http.ResponseWriter, r *http.Request) {

	//fmt.Print(db.Ping())
	userLongitude := r.URL.Query().Get("longitude")
	userLatitude := r.URL.Query().Get("latitude")

	fmt.Fprintf(w, "%s", getDefaultServices(userLatitude,userLongitude))
}

func getServices(uLatitude string,uLongitude string,uService,uRadius string) string {
	ult, err2 := strconv.ParseFloat(uLatitude, 64)
	if err2 != nil {
		fmt.Println("Error")
	}

	ulg, err1 := strconv.ParseFloat(uLongitude, 64)
	if err1 != nil {
		fmt.Println("Error")
	}

	usrRadius, err2 := strconv.ParseFloat(uRadius, 64)
	if err2 != nil {
		fmt.Println("Error")
	}

	//fmt.Println(ulg,ult,usrRadius)

	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	var result string = ""
	rows, err := db.Query("SELECT name,latitude,longitude,couponcode,biztype FROM vendors WHERE biztype=?",uService )
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	//var VendorId Vendor
	var Vendors = make(VendorList,1,100)
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
		log.Fatal(err)
	}
	db.Close()

	jsonInfo,err := json.Marshal(Vendors)
	if err != nil {
		fmt.Println("Error in JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	return S

}

func getDefaultServices(uLatitude string,uLongitude string) string {
	ult, err2 := strconv.ParseFloat(uLatitude, 64)
	if err2 != nil {
		fmt.Println("Error")
	}

	ulg, err1 := strconv.ParseFloat(uLongitude, 64)
	if err1 != nil {
		fmt.Println("Error")
	}

	//fmt.Println(ulg,ult,usrRadius)

	db, err := sql.Open("mysql", "vkonepal:cisco123@/ms")
	checkErr(err)

	var result string = ""
	rows, err := db.Query("SELECT * from vendors" )
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//var VendorId Vendor
	var Vendors = make(VendorList,1,100)
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
		log.Fatal(err)
	}
	db.Close()

	jsonInfo,err := json.Marshal(Vendors)
	if err != nil {
		fmt.Println("Error in JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	return S


	//return result

}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}


func Distance(lat1, lon1, lat2, lon2 float64) float64 {
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
		fmt.Println("Error in JSON marchslling")
	}

	S := string(jsonInfo)
	fmt.Println(S)
	//return S

}


func main() {


	l,err := syslog.New(syslog.LOG_ERR,"VKR")
	defer l.Close()
	if err != nil {
		log.Fatal("Error in the System")
	}else {
		http.HandleFunc("/insertvendor", handler)
		http.HandleFunc("/insertuser", handlerUser)
		http.HandleFunc("/getservices", handlerServices)
		http.HandleFunc("/getdefaultservices", handlerDefaultServices)

		error := http.ListenAndServeTLS(":8443", "/Users/VKONEPAL/IdeaProjects/vkr/server.crt", "/Users/VKONEPAL/IdeaProjects/vkr/server.key", nil)
		if err != nil {
			l.Alert(error.Error())
		}
	}

}
