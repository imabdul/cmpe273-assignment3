# cmpe273-assignment3
 The trip planner is a feature that will take a set of locations from the database and will then check against UBER’s price estimates API to suggest the best possible route in terms of costs and duration.
<pre>
Background:
Starting Location : Fairmont Hotel San Francisco (object_id: 565187c21c4be834a49b77fd)
Location1 : Golden Gate Bridge (object_id: 565187ef1c4be834a49b77fe)
Location2 : Pier 39 (object_id: 5651880c1c4be834a49b77ff)
Location3 : Golden Gate Park (object_id: 565188231c4be834a49b7800)
Location4 : Twin Peaks (object_id: 5651883f1c4be834a49b7802)

1. POST /trips   # Plan a trip

Request Body : {
    "starting_from_location_id": "565187c21c4be834a49b77fd",
    "location_ids" : [ "565187ef1c4be834a49b77fe", "5651880c1c4be834a49b77ff", "565188231c4be834a49b7800", "5651883f1c4be834a49b7802" ] 
}

Response will be like:  {
  "id": "xxxxxxxxxxx", //mongo object id //trip id
  "status": "Planning",
  "starting_from_location_id": "565187c21c4be834a49b77fd",
  "best_route_location_ids": [
    "5651880c1c4be834a49b77ff",
    "565187ef1c4be834a49b77fe",
    "565188231c4be834a49b7800",
    "5651883f1c4be834a49b7802"
  ],
  "total_uber_cost": 64, 
  "total_uber_duration": 4013,
  "total_distance": 23.15
}


2. GET /trips/{trip_id} # Check the trip details and status

Response will be like: {
  "id": "xxxxxxxxxxxxxx",  //mongo object id //trip id
  "status": "Planning",
  "starting_from_location_id": "565187c21c4be834a49b77fd",
  "best_route_location_ids": [
    "5651880c1c4be834a49b77ff",
    "565187ef1c4be834a49b77fe",
    "565188231c4be834a49b7800",
    "5651883f1c4be834a49b7802"
  ],
  "total_uber_cost": 64,
  "total_uber_duration": 4013,
  "total_distance": 23.15
}


3.PUT  /trips/{trip_id}/request # Start the trip by requesting UBER for the first destination. You will call UBER request API to request a car from starting point to the next destination.

Response will be like: {
  "id": "xxxxxxxxx",  //mongo object id //trip id
  "status": "Requesting",
  "starting_from_location_id": "5651880c1c4be834a49b77ff",
  "next_destination_location_id": "565187ef1c4be834a49b77fe",
  "best_route_location_ids": [
    "5651880c1c4be834a49b77ff",
    "565187ef1c4be834a49b77fe",
    "565188231c4be834a49b7800",
    "5651883f1c4be834a49b7802"
  ],
  "total_uber_cost": 64,
  "total_uber_duration": 4013,
  "total_distance": 23.15,
  "uber_wait_time_eta": 0
}
</pre>
