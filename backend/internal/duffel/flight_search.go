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
	Slices     []DuffelSlice     `json:"slices"`
	Passengers []DuffelPassenger `json:"passengers"`
	CabinClass string            `json:"cabin_class"`
}

type DuffelSlice struct {
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
func buildSlices(req DuffelSearchParams) []DuffelSlice {
	if req.TripType == RoundTrip {
		return []DuffelSlice{
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
	return []DuffelSlice{
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
