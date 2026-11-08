package service

type TripTpye string

const (
	OneWay    TripTpye = "one-way"
	RoundTrip TripTpye = "round-trip"
)

type SearchFlightsRequest struct {
	Origin      string
	Destination string
	Departure   string
	Return      string
	Adults      int
	Children    int
	Infants     int
	TripType    TripTpye
}

type SearchFlightsResponse struct {
	FlightOffers []FlightOffer
}

type FlightOffer struct {
	ID            string
	TotalAmount   string
	TotalCurrency string
}
