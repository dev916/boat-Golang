package sites

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"boatfuji.com/api"
)

var (
	btBaseURL          = "https://www.boats.com/"
	sourceBaseURL      = "https://www.boattrader.com"
	btURLPattern       = regexp.MustCompile(`^https://www\.boats\.com/boat/([0-9]*)`)
	btBaseDir          = "/Users/ankitagarwal/ankit_code/boat-golang/harvest/www.boats.com/"
	btBoatMap          = map[string]int64{}
	btBoatsJSONPattern = regexp.MustCompile(`<script>var __REDUX_STATE__=(.*)<\/script>`)
)

func init() {
	api.Sites[btBaseURL] = &Boats{}
}

// Boats accesses https://www.boats.com/
type Boats struct {
	StoreData bool
	WriteSQL  bool
}

// Harvest gets data from the site
func (site *Boats) Harvest(url string) error {
	boatsByFilters = map[string][]string{}
	filtersJSON, err := ioutil.ReadFile(btBaseDir + "filters.json")
	if err == nil {
		err = json.Unmarshal(filtersJSON, &boatsByFilters)
		if err != nil {
			return err
		}
	}
	assetsdir := []string{"boat-sales", "boats"}
	for _, assetdir := range assetsdir {
		os.MkdirAll(btBaseDir+assetdir, 0755)
	}
	switch url {
	case btBaseURL:
		for _, filter := range []string{"type-power", "type-sail", "type-small", "type-pwc"} {
			for page := 1; page < 99999; page++ {
				boatsPage, err := getPage(fmt.Sprintf("%s/boat-sales/%d&%s.html", btBaseDir, page, filter), fmt.Sprintf("%s/boats/%s/page-%d/", sourceBaseURL, filter, page))
				if err != nil {
					return err
				}
				bsBoatIDs := boatsPage.FindN(nil, "//li/@data-listing-id", 0, 9999999, "", "")
				if len(bsBoatIDs) != 0 {
					filtersKey := strings.Split(filter, "-")[1]
					boatsByFilters[filtersKey] = append(bsBoatIDs, boatsByFilters[filter]...)
				} else {
					break
				}
			}
		}
		filtersJSON, _ := json.Marshal(boatsByFilters)
		ioutil.WriteFile(btBaseDir+"filters.json", filtersJSON, 0644)
		return nil
	}
	match := btURLPattern.FindStringSubmatch(url)
	if match == nil {
		return errors.New("BadURL")
	}
	_, err = site.harvestBoat(match[1])
	return err
}

func (site *Boats) harvestBoat(id string) (int64, error) {
	if boatID, ok := btBoatMap[id]; ok {
		return boatID, nil
	}
	url := "https://www.boattrader.com/boat/" + id
	// boat := api.Boat{URLs: []string{url}, Sale: &api.BoatSale{}}
	boatPage, err := getPage(btBaseDir+"boats/"+id+".html", url)
	if err != nil {
		return 0, err
	}
	// check if boat exists on boats.com
	btboatsJSON := boatPage.Find1ByRE(btBoatsJSONPattern, 1, "0", "0")
	re, _ := regexp.Compile(`(<p>.*<\\/p>)`)
	btboatsJSON = re.ReplaceAllString(btboatsJSON, "")
	re, _ = regexp.Compile(`(\,\"dataLayer\".*)`)

	btboatsJSON = re.ReplaceAllString(btboatsJSON, "")
	re, _ = regexp.Compile(`(\}\]\,\"media\")`)
	btboatsJSON = re.ReplaceAllString(btboatsJSON, `}},"media"`)
	re, _ = regexp.Compile(`(\]\}\,\"PremiumPlacementAd\".*)`)
	btboatsJSON = re.ReplaceAllString(btboatsJSON, "")
	var data Document
	if err := json.Unmarshal([]byte(btboatsJSON), &data); err != nil {
		return 0, err
	}
	log.Println(data.App.Data)
	return 0, nil
}

type Document struct {
	App struct {
		Data struct {
			Aliases   Aliases `json:"aliases"`
			Class     string  `json:"class"`
			Condition string  `json:"condition"`
			Contact   struct {
				Address map[string]string `json:"address"`
				Name    string            `json:"name"`
				Phone   string            `json:"phone"`
				Website string            `json:"website"`
			} `json:"contact"`
			Date           Date `json:"date"`
			DealerListings []struct {
				Aliases    Aliases  `json:"aliases"`
				Attributes []string `json:"attributes"`
				BoatName   *string  `json:"boatName,omitempty"`
				Class      string   `json:"class"`
				Classes    []string `json:"classes,omitempty"`
				Condition  string   `json:"condition"`
				Contact    struct {
					Address map[string]string `json:"address"`
					Name    string            `json:"name"`
					Phone   string            `json:"phone"`
				} `json:"contact"`
				Date         Date                `json:"date"`
				Description  *string             `json:"description,omitempty"`
				Descriptions []map[string]string `json:"descriptions"`
				FeatureType  FeatureType         `json:"featureType"`
				Features     struct {
					Covers                 []CodeIDValue `json:"covers"`
					ElectricalEquipment    []CodeIDValue `json:"electricalEquipment"`
					Electronics            []CodeIDValue `json:"electronics"`
					Equipment              Equipment     `json:"equipment"`
					InsideEquipment        []CodeIDValue `json:"insideEquipment"`
					OutsideEquipmentExtras []CodeIDValue `json:"outsideEquipmentExtras"`
					Rigging                []CodeIDValue `json:"rigging"`
					Sails                  []CodeIDValue `json:"sails"`
				} `json:"features"`
				FuelType string `json:"fuelType"`
				Hull     struct {
					Hin      *string `json:"hin,omitempty"`
					KeelType *string `json:"keelType,omitempty"`
					Material string  `json:"material"`
					Shape    *string `json:"shape,omitempty"`
				} `json:"hull"`
				ID    int `json:"id"`
				Legal struct {
					BerthAvailable       *int    `json:"berthAvailable"`
					BuilderName          *string `json:"builderName"`
					CoopTypeID           string  `json:"coopTypeId"`
					DesignerName         *string `json:"designerName"`
					FlagOfRegistry       *string `json:"flagOfRegistry,omitempty"`
					HullWarranty         *string `json:"hullWarranty,omitempty"`
					ListingTypeID        string  `json:"listingTypeId"`
					NotForSaleInUsWaters *int    `json:"notForSaleInUsWaters"`
					TaxCountry           *string `json:"taxCountry,omitempty"`
					TaxStatus            *string `json:"taxStatus,omitempty"`
				} `json:"legal"`
				Location Location `json:"location"`
				Make     string   `json:"make"`
				Media    []Media  `json:"media"`
				Model    string   `json:"model"`
				Owner    struct {
					ID             int            `json:"id"`
					Locale         string         `json:"locale"`
					Location       Location2      `json:"location"`
					Logos          Logos          `json:"logos"`
					Name           string         `json:"name"`
					NormalizedName string         `json:"normalizedName"`
					PhoneNumbers   []PhoneNumbers `json:"phoneNumbers"`
					Type           string         `json:"type"`
				} `json:"owner"`
				PortalLink string `json:"portalLink"`
				Price      Price  `json:"price"`
				Propulsion struct {
					Engines []struct {
						Category          string  `json:"category"`
						DriveType         *string `json:"driveType"`
						FoldingPropeller  bool    `json:"foldingPropeller"`
						Fuel              string  `json:"fuel"`
						Hours             *int    `json:"hours"`
						Make              *string `json:"make,omitempty"`
						Model             *string `json:"model"`
						Power             *Power  `json:"power,omitempty"`
						PropellerMaterial *string `json:"propellerMaterial"`
						PropellerType     *string `json:"propellerType"`
						RopeCutter        bool    `json:"ropeCutter"`
						Year              *int    `json:"year"`
					} `json:"engines"`
				} `json:"propulsion"`
				SalesRep struct {
					FirstName    string         `json:"firstName"`
					ID           int            `json:"id"`
					LastName     string         `json:"lastName"`
					Locale       string         `json:"locale"`
					Location     Location2      `json:"location"`
					PhoneNumbers []PhoneNumbers `json:"phoneNumbers"`
					SalesMessage string         `json:"salesMessage"`
					Type         string         `json:"type"`
				} `json:"salesRep"`
				Specifications struct {
					Accommodation map[string]*int `json:"accommodation"`
					Capacity      struct {
						MaxCapacity   *int `json:"maxCapacity,omitempty"`
						MaxPassengers *int `json:"maxPassengers,omitempty"`
					} `json:"capacity"`
					Dimensions struct {
						Beam    FtM `json:"beam"`
						Lengths struct {
							Nominal   FtM  `json:"nominal"`
							Overall   FtM  `json:"overall"`
							Waterline *FtM `json:"waterline"`
						} `json:"lengths"`
						MaxBridgeClearance *FtM `json:"maxBridgeClearance"`
						MaxDraft           *FtM `json:"maxDraft"`
					} `json:"dimensions"`
					SpeedDistance struct {
						CruisingSpeed *KmhKnMph `json:"cruisingSpeed"`
						MaxHullSpeed  *KmhKnMph `json:"maxHullSpeed"`
						Range         *struct {
							Km  int     `json:"km"`
							Mi  float64 `json:"mi"`
							Nmi int     `json:"nmi"`
						} `json:"range"`
					} `json:"speedDistance"`
					Weights struct {
						Displacement *D `json:"displacement"`
						Dry          *D `json:"dry"`
					} `json:"weights"`
				} `json:"specifications"`
				Status string `json:"status"`
				Tanks  struct {
					FreshWater []CapacityQuantityTankMaterial `json:"freshWater"`
					Fuel       []CapacityQuantityTankMaterial `json:"fuel"`
					Holding    []CapacityQuantityTankMaterial `json:"holding"`
				} `json:"tanks"`
				Title *string `json:"title,omitempty"`
				Type  string  `json:"type"`
				Year  int     `json:"year"`
			} `json:"dealerListings"`
			Descriptions []struct {
				Description     *string `json:"description,omitempty"`
				DescriptionType string  `json:"descriptionType"`
				Title           *string `json:"title,omitempty"`
				Visibility      *string `json:"visibility,omitempty"`
			} `json:"descriptions"`
			FeatureType FeatureType `json:"featureType"`
			Features    struct {
				Equipment Equipment `json:"equipment"`
			} `json:"features"`
			FuelType string `json:"fuelType"`
			Hull     struct {
				Material string `json:"material"`
			} `json:"hull"`
			ID    int `json:"id"`
			Legal struct {
				CoopTypeID    string `json:"coopTypeId"`
				ListingTypeID string `json:"listingTypeId"`
			} `json:"legal"`
			Location Location `json:"location"`
			Make     string   `json:"make"`
			Model    string   `json:"model"`
			Owner    struct {
				BoattraderID   string    `json:"boattraderId"`
				ID             int       `json:"id"`
				Locale         string    `json:"locale"`
				Location       Location2 `json:"location"`
				Logos          Logos     `json:"logos"`
				Name           string    `json:"name"`
				NormalizedName string    `json:"normalizedName"`
				Type           string    `json:"type"`
			} `json:"owner"`
			PortalLink string `json:"portalLink"`
			Price      Price  `json:"price"`
			Propulsion struct {
				Engines []struct {
					Category          string `json:"category"`
					FoldingPropeller  bool   `json:"foldingPropeller"`
					Fuel              string `json:"fuel"`
					Make              string `json:"make"`
					Model             string `json:"model"`
					Power             Power  `json:"power"`
					PropellerMaterial string `json:"propellerMaterial"`
					RopeCutter        bool   `json:"ropeCutter"`
				} `json:"engines"`
			} `json:"propulsion"`
			SalesRep struct {
				FirstName    string    `json:"firstName"`
				ID           int       `json:"id"`
				LastName     string    `json:"lastName"`
				Locale       string    `json:"locale"`
				Location     Location2 `json:"location"`
				SalesMessage string    `json:"salesMessage"`
				Type         string    `json:"type"`
			} `json:"salesRep"`
			Specifications struct {
				Accommodation struct {
					Heads int `json:"heads"`
				} `json:"accommodation"`
				Capacity struct {
				} `json:"capacity"`
				Dimensions struct {
					Beam    FtM `json:"beam"`
					Lengths struct {
						Nominal struct {
							Ft int     `json:"ft"`
							M  float64 `json:"m"`
						} `json:"nominal"`
						Overall FtM `json:"overall"`
					} `json:"lengths"`
				} `json:"dimensions"`
				SpeedDistance struct {
				} `json:"speedDistance"`
				Weights struct {
				} `json:"weights"`
			} `json:"specifications"`
			Status string `json:"status"`
			Tanks  struct {
			} `json:"tanks"`
			Title string `json:"title"`
			Type  string `json:"type"`
			Year  int    `json:"year"`
		} `json:"data"`
		Errors    bool   `json:"errors"`
		IsWorking bool   `json:"isWorking"`
		Message   string `json:"message"`
		Params    struct {
			ID string `json:"id"`
		} `json:"params"`
		Success bool `json:"success"`
	} `json:"app"`
}
type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}
type Aliases struct {
	Bcna       string `json:"bcna"`
	Boattrader string `json:"boat-trader"`
	Imt        int    `json:"imt"`
	Yachtworld string `json:"yachtworld"`
}
type FtM struct {
	Ft float64 `json:"ft"`
	M  float64 `json:"m"`
}
type Capacity struct {
	Gal   int     `json:"gal"`
	Galuk float64 `json:"galuk"`
	L     float64 `json:"l"`
}
type Codes struct {
	IsoAlpha2      string `json:"isoAlpha2"`
	IsoSubdivision string `json:"isoSubdivision"`
}
type CodeIDValue []struct {
	Code  string  `json:"code"`
	ID    string  `json:"id"`
	Value *string `json:"value,omitempty"`
}
type KmhKnMph struct {
	Kmh float64 `json:"kmh"`
	Kn  int     `json:"kn"`
	Mph float64 `json:"mph"`
	Rpm *int    `json:"rpm"`
}
type Date struct {
	Created  string `json:"created"`
	Modified string `json:"modified"`
}
type D struct {
	Kg float64 `json:"kg"`
	Lb int     `json:"lb"`
}
type Equipment struct {
	JoystickControl int `json:"joystickControl"`
	TrimTab         int `json:"trimTab"`
}
type FeatureType struct {
	Enhanced  bool `json:"enhanced"`
	Sponsored bool `json:"sponsored"`
}
type CapacityQuantityTankMaterial []struct {
	Capacity     *Capacity `json:"capacity,omitempty"`
	Quantity     *int      `json:"quantity"`
	TankMaterial *string   `json:"tankMaterial"`
}
type Location struct {
	Address     Address   `json:"address"`
	Codes       Codes     `json:"codes"`
	Coordinates []float64 `json:"coordinates"`
}
type Logos struct {
	Default  string `json:"default"`
	Enhanced string `json:"enhanced"`
}
type Media []struct {
	Date         Date   `json:"date"`
	Format       string `json:"format"`
	Height       int    `json:"height"`
	LanguageCode string `json:"languageCode"`
	MediaType    string `json:"mediaType"`
	SortOrder    int    `json:"sortOrder"`
	Status       string `json:"status"`
	URL          string `json:"url"`
	Width        int    `json:"width"`
}

type PhoneNumbers []struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}
type Power struct {
	Hp int     `json:"hp"`
	KW float64 `json:"kW"`
}
type Location2 struct {
	Address map[string]string `json:"address"`
	Codes   Codes             `json:"codes"`
}
type Type struct {
	Amount        map[string]float64 `json:"amount"`
	BaseAmount    map[string]float64 `json:"baseAmount"`
	SpecialAmount map[string]float64 `json:"specialAmount"`
}
type Price struct {
	EnteredCurrency string `json:"enteredCurrency"`
	Hidden          int    `json:"hidden"`
	Type            Type   `json:"type"`
}
