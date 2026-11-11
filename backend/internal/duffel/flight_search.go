package duffel

type TripType string

const (
	OneWay    TripType = "one-way"
	RoundTrip TripType = "round-trip"
)

type DuffelSearchParams struct {
	Origin      string
	Destination string
	Departure   string
	Return      string
	Adults      int
	Children    int
	Infants     int
	CabinClass  string
	TripType    TripType
}

type DuffelSearchRequest struct {
	Data DuffelSearchData `json:"data"`
}

type DuffelSearchData struct {
	Slices     []DuffelSearchSlice     `json:"slices"`
	Passengers []DuffelPassenger `json:"passengers"`
	CabinClass string            `json:"cabin_class"`
}

type DuffelOffer struct {
	ID            string `json:"id"`
	TotalAmount   string `json:"total_amount"`
	TotalCurrency string `json:"total_currency"`
	Slices        []DuffelOfferSlice   `json:"slices"`
}

type DuffelSegment struct {
	ID 		string        `json:"id"`
	Duration   string `json:"duration"`
	MarketingCarrier DuffelCarrier `json:"marketing_carrier"`
	Aircraft	  DuffelAircraft `json:"aircraft"`
	Origin      DuffelAirport `json:"origin"`
	Destination DuffelAirport `json:"destination"`
	Departure   string `json:"departing_at"`
	Arrival     string `json:"arriving_at"`
}

type DuffelOfferSlice struct {
    Origin      DuffelAirport   `json:"origin"`
    Destination DuffelAirport  `json:"destination"`
    Segments    []DuffelSegment `json:"segments"`
}

type DuffelAirport struct {
	IATA string `json:"iata_code"`
	Name string `json:"name"`
}

type DuffelCarrier struct {
	IATA string `json:"iata_code"`
	Name string `json:"name"`
}

type DuffelAircraft struct {
	IATA string `json:"iata_code"`
	Name string `json:"name"`
}

type DuffelSearchSlice struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
}

type DuffelPassenger struct {
	Type string `json:"type"`
}

func buildPassengers(adults, children, infants int) []DuffelPassenger {
	passengers := make([]DuffelPassenger, 0, adults+children+infants)
	for i := 0; i < adults; i++ {
		passengers = append(passengers, DuffelPassenger{Type: "adult"})
	}
	for i := 0; i < children; i++ {
		passengers = append(passengers, DuffelPassenger{Type: "child"})
	}
	for i := 0; i < infants; i++ {
		passengers = append(passengers, DuffelPassenger{Type: "infant"})
	}

	return passengers
}
func buildSlices(req DuffelSearchParams) []DuffelSearchSlice {
	if req.TripType == RoundTrip {
		return []DuffelSearchSlice{
			{
				Origin:        req.Origin,
				Destination:   req.Destination,
				DepartureDate: req.Departure,
			},
			{
				Origin:        req.Destination,
				Destination:   req.Origin,
				DepartureDate: req.Return,
			},
		}
	}
	return []DuffelSearchSlice{
		{
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.Departure,
		},
	}
}
func buildSearchRequest(req DuffelSearchParams) DuffelSearchRequest {
	return DuffelSearchRequest{
		Data: DuffelSearchData{
			Slices:     buildSlices(req),
			Passengers: buildPassengers(req.Adults, req.Children, req.Infants),
			CabinClass: req.CabinClass,
		},
	}
}
