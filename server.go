package main

import (

	//third party modules
	"github.com/jmoiron/jsonq"
	"github.com/r-medina/go-uber"
	"github.com/julienschmidt/httprouter"

	//Go internal modules
	"fmt"
    "sort"
    "errors"
    "strconv"
    "strings"
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"

	//mongo specific modules
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

)


type ETA struct {
	Eta                              int                    `json:"eta"`
	RequestID                        string                 `json:"request_id"`
	Status                           string                 `json:"status"`
	SurgeMultiplier                  float64                `json:"surge_multiplier"`
}

const (
SERVER_TOKEN = "kZv1GJpCYc00e21Kc1GjnXixWoE-p0J-uy00xdnP"
)

type dataSlice []Data
var originId string
var locationsArray []string
var nextId string

//var compare string

/*
type trip struct {
	id 									bson.ObjectId 						`json:"id" bson:"_id"`
	trip_staus 							string  							`json:"status" bson:"status"`
	start_location_id 					string 								`json:"starting_from_location_id" bson:"starting_from_location_id"`
	location_ids_array		 			[]string 							`json:"best_route_location_ids" bson:"best_route_location_ids"`
	total_costs							int 								`json:"total_uber_cost" bson:"total_uber_cost"`
	total_duration 						int 								`json:"total_uber_duration" bson:"total_uber_duration"`
	distance 							float64 							`json:"total_distance" bson:"total_distance"`
}
*/


type tripDetails struct {
Id 									bson.ObjectId 						`json:"id" bson:"_id"`
Status 								string  							`json:"status" bson:"status"`
Starting_from_location_id 			string 								`json:"starting_from_location_id" bson:"starting_from_location_id"`
 Best_route_location_ids 			[]string 							`json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs 					int 								`json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration 				int 								`json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance 					float64 							`json:"total_distance" bson:"total_distance"`
}


type tripDetailsStatus struct {
Id 									bson.ObjectId 						`json:"id" bson:"_id"`
Status 								string  							`json:"status" bson:"status"`
Starting_from_location_id 			string 								`json:"starting_from_location_id" bson:"starting_from_location_id"`
 Next_destination_location_id 		string 								`json:"next_destination_location_id" bson:"next_destination_location_id"`
 Best_route_location_ids	 		[]string 							`json:"best_route_location_ids" bson:"best_route_location_ids"`
  Total_uber_costs 					int 								`json:"total_uber_cost" bson:"total_uber_cost"`
  Total_uber_duration 				int 								`json:"total_uber_duration" bson:"total_uber_duration"`
  Total_distance 					float64 							`json:"total_distance" bson:"total_distance"`
  Uber_wait_time_eta 				int 								`json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`
}


type Data struct{
			id 						string
			price 					int
			duration 				int
			distance 				float64
}

type crdnt struct {
			lat 					float64
			lng 					float64
}

type tripsRequest struct {
    		LocationIds           	[]string `json:"location_ids"`
    		StartingFromLocationID 	string   `json:"starting_from_location_id"`
}

type locationDetails struct {
   			 Id 							bson.ObjectId 					`json:"id" bson:"_id"`
			 Name 							string 							`json:"name" bson:"name"`
			 Address						string 							`json:"address" bson:"address"`
			 City 							string 							`json:"city" bson:"city"`
			 State							string 							`json:"state" bson:"state"`
			 Zip 							string 							`json:"zip" bson:"zip"`


    Coordinate struct {
			Lat 							float64 						`json:"lat" bson:"lat"`
			Lng 							float64 						`json:"lng" bson:"lng"`
    } 																		`json:"coordinate" bson:"coordinate"`
}

// Method to create locations
func createLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u locationDetails
    locationURL := "http://maps.google.com/maps/api/geocode/json?address="
    json.NewDecoder(req.Body).Decode(&u)

    //Random Id generation
    u.Id = bson.NewObjectId()

    locationURL = locationURL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    locationURL = strings.Replace(locationURL, " ", "+", -1)
    fmt.Println("Location URL :"+ locationURL)

    //Google Maps API call here
    response, err := http.Get(locationURL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println("Status:" + status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
   if err != nil {
       fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = lat
    u.Coordinate.Lng = lng

    //Data Persistence in MongoDB
    newSession().DB("cmpe273").C("locations").Insert(u)

    reply, _ := json.Marshal(u)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}

//Delete a Location from locations collection
func deleteLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	locId :=  p.ByName("location_id")

	if !bson.IsObjectIdHex(locId) {
		rw.WriteHeader(404)
		return
	}

	dataid := bson.ObjectIdHex(locId)

	// User gets deleted here
	if err := newSession().DB("cmpe273").C("locations").RemoveId(dataid); err != nil {
		rw.WriteHeader(404)
		return
	}

	rw.WriteHeader(200)
}

//Method to retrieve locations
func getLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    locId :=  p.ByName("location_id")
//	fmt.Println("ObjectId: " + p.ByName("location_id") )

    if !bson.IsObjectIdHex(locId) {
        rw.WriteHeader(404)
        return
    }

    dataId := bson.ObjectIdHex(locId)

    responseObj := locationDetails{}

    if err := newSession().DB("cmpe273").C("locations").FindId(dataId).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}

//Update Location 
func updateLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var updLocation locationDetails
	locId :=  p.ByName("location_id")

    URL := "http://maps.google.com/maps/api/geocode/json?address="

    //transfer the data into the local object
    json.NewDecoder(req.Body).Decode(&updLocation)

    URL = URL + updLocation.Address+ " " + updLocation.City + " " + updLocation.State + " " + updLocation.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("Location URL: "+ URL)

    //Google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    fmt.Println("Status: " + status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    lat, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    lng, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    updLocation.Coordinate.Lat = lat
    updLocation.Coordinate.Lng = lng

    dataid := bson.ObjectIdHex(locId)
    var data = locationDetails{
        Address: updLocation.Address,
        City: updLocation.City,
        State: updLocation.State,
        Zip: updLocation.Zip,
    }
    //updatedata
    fmt.Println(data)
    //store data
    newSession().DB("cmpe273").C("locations").Update(bson.M{"_id":dataid }, bson.M{"$set": bson.M{ "address": updLocation.Address,
        "city": updLocation.City, "state": updLocation.State,"zip": updLocation.Zip, "coordinate.lat":updLocation.Coordinate.Lat, "coordinate.lng":updLocation.Coordinate.Lng}})

    responseObj := locationDetails{}

    //retrive the response data
    if err := newSession().DB("cmpe273").C("locations").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    // interface into JSON struct
    reply, _ := json.Marshal(responseObj)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)

}



//Session creation of MongoDB
func newSession() *mgo.Session {
    //Connecting to Mongo Lab
    s, err := mgo.Dial("mongodb://imabdul:imabdul@ds043694.mongolab.com:43694/cmpe273")

    // Check if mongo server running
    if err != nil {
        panic(err)
    }
    return s
}


//function to get coordinates of objectid
func getDetails(x string) (y crdnt) {
    responseObj := locationDetails{}
//dataid := bson.ObjectId(x)
    if err := newSession().DB("cmpe273").C("locations").Find(bson.M{"_id": bson.ObjectIdHex(x)}).One(&responseObj); err != nil {
        z := crdnt{}
    return z
}
    p := crdnt{
    lat: responseObj.Coordinate.Lat,
    lng: responseObj.Coordinate.Lng,
    }
    return p
    
}


//gives estimated price from last location to origin location
func getpricetostart(x string)(y Data){
	var price []int
	response, err := http.Get(x)
	if err != nil {
		return
	}
	defer response.Body.Close()
	resp := make(map[string]interface{})
	body, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return
	}
	ptr := resp["prices"].([]interface{})
	jq := jsonq.NewQuery(resp)
	for i, _ := range ptr {
		pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
		price = append(price,pr)
	}
	min := price[0]
	for j, _ := range price {
		if(price[j]<=min && price[j]!=0){
			min = price[j]
		}
	}
	du,_:=jq.Int("prices","0","duration")
	dist,_:=jq.Float("prices","0","distance")
	d := Data{
		id:"",
		price : min,
		duration:du,
		distance:dist,
	}
	return d
}


//function to get estimated price from uber API
func getEstimatedPrice(x string, z string)(y Data){
response, err := http.Get(x)
    if err != nil {
        return
    }
    defer response.Body.Close()
    var price []int
    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
	panic(err)
        return
    }
    ptr := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)
     for i, _ := range ptr {
     pr,_ := jq.Int("prices",fmt.Sprintf("%d", i),"low_estimate")
     price = append(price,pr)
	}
     min := price[0]
     for j, _ := range price {
     if(price[j]<=min && price[j]!=0){
     min = price[j]
     }
     }
     du,_:=jq.Int("prices","0","duration")
     dist,_:=jq.Float("prices","0","distance")
     data := Data{
     id:z,
     price:min,
     duration:du,
     distance:dist,
     }
    return data     
}


//its a sort.Interface utility
func (d dataSlice) Len() int {
	return len(d)
}

//its a sort.Interface utility
func (d dataSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

//its a sort.Interface utility.
func (d dataSlice) Less(i, j int) bool {
	return d[i].price < d[j].price 
}
//In a given set of locations, finds least price among them
func prioritize(x map[string]Data)(y Data) {
	m := x
	s := make(dataSlice, 0, len(m))
	for _, d := range m {
		s = append(s, d)
	}		
	sort.Sort(s)
	return s[0]
}

//To delete Id from the list
func deleteId(s []string, p string)(x []string) {
    var r []string
    for _, str := range s {
        if str != p {
            r = append(r, str)
        }
    }
    return r
}


//To sum int
func Sumint(a []int) (sum int) {
    for _, v := range a {
        sum += v
    }
    return
}

//To sum float
func Sumfloat(a []float64) (sum float64) {
	for _, v := range a {
		sum += v
	}
	return
}

//handles trip request
func optimalPathRequest(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
    decoder := json.NewDecoder(req.Body)
    var tripsReq tripsRequest
    err := decoder.Decode(&tripsReq)
    if err != nil {
        panic(err)
    }
    Start := tripsReq.StartingFromLocationID
    locIds := tripsReq.LocationIds
    var T tripDetails
    var z crdnt
    var tp []int
    var td []float64
    var tdu []int

   for arraylength:=len(locIds); arraylength>0; arraylength--{
    z = getDetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := []crdnt{}
    for i := 0; i < len(locIds); i++ {
       y := getDetails(locIds[i])
       x = append(x,y)
   }
   tdata := map[string]Data{}
      for i:=0;i<len(x);i++{
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=tyC8DGEaUgBO68yd4rE9RIKF4PyweeZq0uH-bz9-",start_lat,start_lng,x[i].lat,x[i].lng)
      d:= getEstimatedPrice(url, locIds[i])
      tdata[locIds[i]] = d
      }
   da:=prioritize(tdata)
  T.Best_route_location_ids = append(T.Best_route_location_ids,da.id)
   tp = append(tp,da.price)
   td = append(td,da.distance)
   tdu = append(tdu,da.duration)
   locIds = deleteId(locIds,da.id)
   Start=da.id
   }
   if(locIds ==nil){
   z = getDetails(Start)
    start_lat := z.lat
    start_lng := z.lng
    x := crdnt{}
    y := getDetails(tripsReq.StartingFromLocationID)
    x.lat=y.lat
    x.lng=y.lng
       tripdata := map[string]Data{}
      url := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&server_token=tyC8DGEaUgBO68yd4rE9RIKF4PyweeZq0uH-bz9-",start_lat,start_lng,x.lat,x.lng)
      d:=getpricetostart(url)
      tripdata[Start] = d
   tp = append(tp,d.price)
   td = append(td,d.distance)
   tdu = append(tdu,d.duration)
   }
   

T.Id = bson.NewObjectId()
T.Status = "Planning"
T.Starting_from_location_id= tripsReq.StartingFromLocationID
 T.Best_route_location_ids = T.Best_route_location_ids
T.Total_uber_costs = Sumint(tp)
T.Total_uber_duration = Sumint(tdu)
T.Total_distance = Sumfloat(td) 
newSession().DB("cmpe273").C("tripdetails").Insert(T)
    // interface to JSON struct
    reply, _ := json.Marshal(T)
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", reply)
        
    }



//gets you the trip ETA
func getTripETA(x float64,y float64,z string)(p int){
	lat := strconv.FormatFloat(x, 'E', -1, 64)
	lng := strconv.FormatFloat(y, 'E', -1, 64)
	url := "https://sandbox-api.uber.com/v1/requests"
	var jsonStr = []byte(`{
"start_latitude":"`+lat+`",
"start_longitude":"`+lng+`",
"product_id":"`+z+`",
}`)
	//fmt.Println( jsonStr)
	//fmt.Println( lat)
	//fmt.Println( lng)
	//fmt.Println( p)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsImhpc3RvcnkiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiYWE4ZDc5OGUtNzk2YS00OGE2LWFhOWYtOGIwNTk4MjUyYWFhIiwiaXNzIjoidWJlci11czEiLCJqdGkiOiJhMWE5MTQxZi1jNDk2LTRhNjItOTdhNi1lZDY4ZWZiZWFmMTYiLCJleHAiOjE0NTA4NTM1MzcsImlhdCI6MTQ0ODI2MTUzNywidWFjdCI6Ik4xdUFjdXpzeDR4dFdOaEtDS2FHZXpaV3hZdElXcSIsIm5iZiI6MTQ0ODI2MTQ0NywiYXVkIjoicW01MEZFNnd5S194N3BGbkNlNHFFOU1JMTFQeGczQmgifQ.Pq-5bLxfiwZ23T6e2XoDbvGRKfPHmdzpts-zC7r4m--eu095jLsElpbACfgSZXrs0IBP-4rCCegw5_SUUTJN39QffKVO1P2OCqCICFQKwmXCe6QsZY3rayB6HuIgV5ajgY8wGwOSZ19X2gvPj138hpj3LI_4TJM4Ff0bIW0aLYigJfc2l8544tt3CBPD20Ax_pAJzUrseZHXfSmlGdmcHhqIVPr7cH6hpBeGKIWNnJkuE7IJ01Xvtc9XIvffJPjRv-NiFbjovJQ9ra546cNaUq7kHAFb8vnH9YcF0uTvu8gziz2d-ORgAavkg2-KanljXZI387eOlMlNqIB6I2SrRw")
	req.Header.Set("Content-Type", "application/json")
	var resp1 ETA
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body,&resp1)
	if err != nil {
		panic(err)
	}

	rid:= resp1.Eta
	//fmt.Println(resp1)
	return rid

}


//fuction to handle PUT /trips/:trip_id/request
func tripStatusRequest(rw http.ResponseWriter, req *http.Request, p httprouter.Params){
	client := uber.NewClient(SERVER_TOKEN)
	tripId :=  p.ByName("trip_id")
	if !bson.IsObjectIdHex(tripId) {
		rw.WriteHeader(404)
		return
	}

	dataId := bson.ObjectIdHex(tripId)

	responseObj := tripDetails{}

	if err := newSession().DB("cmpe273").C("tripdetails").FindId(dataId).One(&responseObj); err != nil {
		rw.WriteHeader(404)
		return
	}
	if(nextId ==""){
		originId =responseObj.Starting_from_location_id
		locationsArray =responseObj. Best_route_location_ids
		z := getDetails(responseObj.Starting_from_location_id)
		start_lat := z.lat
		start_lng := z.lng
		products,_ := client.GetProducts(start_lat,start_lng)
		productid := products[0].ProductID
		eta:= getTripETA(start_lat,start_lng,productid)
		nextId = locationsArray[0]
		reply := tripDetailsStatus{
			Id:responseObj.Id,
			Starting_from_location_id :originId,
			Best_route_location_ids:responseObj. Best_route_location_ids,
			Total_uber_costs:responseObj.Total_uber_costs,
			Total_uber_duration:responseObj.Total_uber_duration,
			Total_distance:responseObj.Total_distance,
			Uber_wait_time_eta: eta,
			Status : "Requesting",
			Next_destination_location_id: nextId,
		}
		newSession().DB("cmpe273").C("tripdetails").Update(bson.M{"_id":dataId }, bson.M{"$set": bson.M{ "status": "Requesting"}})
		originId = nextId
		locationsArray = deleteId(locationsArray, nextId)
		if(locationsArray !=nil){
			nextId = locationsArray[0]
		}else{
			nextId = "empty"
		}
		res, _ := json.Marshal(reply)
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(rw, "%s", res)
	}else if(locationsArray !=nil){
		if(nextId !="empty"){
			z := getDetails(originId)
			start_lat := z.lat
			start_lng := z.lng
			products,_ := client.GetProducts(start_lat,start_lng)
			productid := products[0].ProductID
			tripETA := getTripETA(start_lat,start_lng,productid)
			reply := tripDetailsStatus{
				Id:responseObj.Id,
				Starting_from_location_id :originId,
				Best_route_location_ids:responseObj. Best_route_location_ids,
				Total_uber_costs:responseObj.Total_uber_costs,
				Total_uber_duration:responseObj.Total_uber_duration,
				Total_distance:responseObj.Total_distance,
				Uber_wait_time_eta: tripETA,
				Status : "Requesting",
				Next_destination_location_id: nextId,
			}
			newSession().DB("cmpe273").C("tripdetails").Update(bson.M{"_id":dataId }, bson.M{"$set": bson.M{ "status": "Requesting"}})
			originId = nextId
			locationsArray = deleteId(locationsArray, nextId)
			if(locationsArray !=nil){
				nextId = locationsArray[0]
			}else{
				nextId = "empty"
			}
			res, _ := json.Marshal(reply)
			rw.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(rw, "%s", res)
		}
	}else if(nextId =="empty"){
		z := getDetails(originId)
		start_lat := z.lat
		start_lng := z.lng
		products,_ := client.GetProducts(start_lat,start_lng)
		productid := products[0].ProductID
		eta:= getTripETA(start_lat,start_lng,productid)
		reply := tripDetailsStatus{
			Id:responseObj.Id,
			Starting_from_location_id :originId,
			Best_route_location_ids:responseObj. Best_route_location_ids,
			Total_uber_costs:responseObj.Total_uber_costs,
			Total_uber_duration:responseObj.Total_uber_duration,
			Total_distance:responseObj.Total_distance,
			Uber_wait_time_eta: eta,
			Status : "Requesting",
			Next_destination_location_id: responseObj.Starting_from_location_id,
		}
		newSession().DB("cmpe273").C("tripdetails").Update(bson.M{"_id":dataId }, bson.M{"$set": bson.M{ "status": "Requesting"}})
		nextId ="complete"
		res, _ := json.Marshal(reply)
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(rw, "%s", res)
	}else{
		reply := tripDetailsStatus{
			Id:responseObj.Id,
			Starting_from_location_id :responseObj.Starting_from_location_id,
			Best_route_location_ids:responseObj. Best_route_location_ids,
			Total_uber_costs:responseObj.Total_uber_costs,
			Total_uber_duration:responseObj.Total_uber_duration,
			Total_distance:responseObj.Total_distance,
			Uber_wait_time_eta: 0 ,
			Status : "Finished",
			Next_destination_location_id: "",
		}
		newSession().DB("cmpe273").C("tripdetails").Update(bson.M{"_id":dataId }, bson.M{"$set": bson.M{ "status": "Finished"}})
		nextId =""
		res, _ := json.Marshal(reply)
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(rw, "%s", res)
	}
}



//Retrieve trip based on the trip id
func getTripDetails(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    tripId :=  p.ByName("trip_id")

    if !bson.IsObjectIdHex(tripId) {
        rw.WriteHeader(404)
        return
    }

    dataid := bson.ObjectIdHex(tripId)

    responseObj := tripDetails{}

    if err := newSession().DB("cmpe273").C("tripdetails").FindId(dataid).One(&responseObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    reply, _ := json.Marshal(responseObj)

    
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", reply)
}
  
    


func main()  {

    router := httprouter.New()

	router.GET("/locations/:location_id", getLocations)

	router.POST("/locations", createLocations)

	router.DELETE("/locations/:location_id", deleteLocations)

    router.PUT("/locations/:location_id", updateLocations)

    router.POST("/trips", optimalPathRequest)

   router.PUT("/trips/:trip_id/request", tripStatusRequest)

	router.GET("/trips/:trip_id", getTripDetails)

	server := http.Server{
		Addr:        "0.0.0.0:8090",
		Handler: router,
    }
    server.ListenAndServe()
}