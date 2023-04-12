package location

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	locationSdk "github.com/aws/aws-sdk-go/service/locationservice"
)

type LocationConfig struct {
	Client    *locationSdk.LocationService
	IndexName string
}

type Index struct {
	IndexName string
}

type PlaceResult struct {
	Place Place `json:"place"`
}

type Place struct {
	Country      string   `json:"country"`
	Geometry     Geometry `json:"geometry"`
	Label        string   `json:"label"`
	Municipality string   `json:"municipality"`
	Neighborhood string   `json:"neighborhood"`
	PostalCode   string   `json:"postalCode"`
	Region       string   `json:"region"`
	Street       string   `json:"Street"`
	SubRegion    string   `json:"subRegion"`
}

type Geometry struct {
	Point [2]float64 `json:"point"`
}

func Init(indexName string) LocationConfig {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return LocationConfig{
		Client:    locationSdk.New(sess),
		IndexName: indexName,
	}
}

func (loc LocationConfig) FindOrCreateIndex() error {
	log.Printf("[CORE][LOCATION] FindOrCreateIndex Index %v ", loc.IndexName)
	index, err := loc.RetreiveIndex()
	if err != nil && index.IndexName == "" {
		log.Printf("[CORE][LOCATION] Index %+v does not exist, creating it", loc.IndexName)
		inpt := &locationSdk.CreatePlaceIndexInput{
			DataSource:  aws.String("Esri"),
			IndexName:   aws.String(loc.IndexName),
			PricingPlan: aws.String("RequestBasedUsage"),
		}
		_, err := loc.Client.CreatePlaceIndex(inpt)
		if err != nil {
			fmt.Printf("[CORE][LOCATION] Error creating index: %v", err)
			return err
		}
	}
	fmt.Printf("[CORE][LOCATION] Index %+v already exist", index.IndexName)
	return nil
}

func (loc LocationConfig) RetreiveIndex() (Index, error) {
	input := &locationSdk.DescribePlaceIndexInput{
		IndexName: aws.String(loc.IndexName),
	}
	res, err := loc.Client.DescribePlaceIndex(input)
	if err != nil {
		log.Printf("[CORE][LOCATION] Error Retreiving index: %v", err)
		return Index{}, err
	}
	return Index{IndexName: *res.IndexName}, nil
}

func (loc LocationConfig) DeleteIndex() error {
	log.Printf("[CORE][LOCATION] Deleting Index %+v ", loc.IndexName)
	input := &locationSdk.DeletePlaceIndexInput{
		IndexName: aws.String(loc.IndexName),
	}
	_, err := loc.Client.DeletePlaceIndex(input)
	return err
}

func (loc LocationConfig) ReverseGeocode(lat float64, long float64) ([]PlaceResult, error) {
	log.Printf("[CORE][LOCATION] Search Index %+v with query (%v,%v)", loc.IndexName, lat, long)
	input := &locationSdk.SearchPlaceIndexForPositionInput{
		IndexName: aws.String(loc.IndexName),
		//careful to the order=> long, lat
		Position: []*float64{aws.Float64(long), aws.Float64(lat)},
	}
	response, err := loc.Client.SearchPlaceIndexForPosition(input)
	if err != nil {
		log.Printf("[CORE][LOCATION] Error while searching %v", err)
		return []PlaceResult{}, err
	}
	places := []PlaceResult{}
	for _, res := range response.Results {
		if res.Place != nil {
			var label, municipality, neighborhood, postalCode, region, street, subRegion string
			//var distance float64
			if res.Place.Label != nil {
				label = *res.Place.Label
			}
			if res.Place.Municipality != nil {
				municipality = *res.Place.Municipality
			}
			if res.Place.Neighborhood != nil {
				neighborhood = *res.Place.Neighborhood
			}
			if res.Place.PostalCode != nil {
				postalCode = *res.Place.PostalCode
			}
			if res.Place.Region != nil {
				region = *res.Place.Region
			}
			if res.Place.Street != nil {
				street = *res.Place.Street
			}
			if res.Place.SubRegion != nil {
				subRegion = *res.Place.SubRegion
			}
			/*if res.Distance != nil {
				distance = *res.Distance
			}*/
			places = append(places, PlaceResult{
				Place: Place{
					Country: *res.Place.Country,
					Geometry: Geometry{
						Point: [2]float64{*res.Place.Geometry.Point[0], *res.Place.Geometry.Point[1]},
					},
					Label:        label,
					Municipality: municipality,
					Neighborhood: neighborhood,
					PostalCode:   postalCode,
					Region:       region,
					Street:       street,
					SubRegion:    subRegion,
				},
			})
		}

	}
	log.Printf("[CORE][LOCATION] Search for position result: %+v", places)
	return places, nil
}

func (loc LocationConfig) Search(text string) ([]PlaceResult, error) {
	log.Printf("[CORE][LOCATION] Search Index %+v with query %v", loc.IndexName, text)
	input := &locationSdk.SearchPlaceIndexForTextInput{
		IndexName: aws.String(loc.IndexName),
		Text:      aws.String(text),
	}
	response, err := loc.Client.SearchPlaceIndexForText(input)
	if err != nil {
		log.Printf("[CORE][LOCATION] Error while searching %v", err)
		return []PlaceResult{}, err
	}
	places := []PlaceResult{}
	for _, res := range response.Results {
		if res.Place != nil {
			var label, municipality, neighborhood, postalCode, region, street, subRegion string
			if res.Place.Label != nil {
				label = *res.Place.Label
			}
			if res.Place.Municipality != nil {
				municipality = *res.Place.Municipality
			}
			if res.Place.Neighborhood != nil {
				neighborhood = *res.Place.Neighborhood
			}
			if res.Place.PostalCode != nil {
				postalCode = *res.Place.PostalCode
			}
			if res.Place.Region != nil {
				region = *res.Place.Region
			}
			if res.Place.Street != nil {
				street = *res.Place.Street
			}
			if res.Place.SubRegion != nil {
				subRegion = *res.Place.SubRegion
			}
			places = append(places, PlaceResult{
				Place: Place{
					Country: *res.Place.Country,
					Geometry: Geometry{
						Point: [2]float64{*res.Place.Geometry.Point[0], *res.Place.Geometry.Point[1]},
					},
					Label:        label,
					Municipality: municipality,
					Neighborhood: neighborhood,
					PostalCode:   postalCode,
					Region:       region,
					Street:       street,
					SubRegion:    subRegion,
				},
			})
		}

	}
	return places, nil
}
