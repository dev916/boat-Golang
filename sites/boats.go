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
	sourceBaseURL      = "https://www.boattrader.com/"
	btURLPattern       = regexp.MustCompile(`^https://www\.boats\.com/boat/([0-9]*)`)
	btBaseDir          = "harvest/www.boats.com/"
	btBoatMap          = map[string]int64{}
	btBoatsJSONPattern = regexp.MustCompile(`<script>var __REDUX_STATE__=(.*)<\/script>`)
	boatTypePattern    = regexp.MustCompile(`"type":"((.*?))"`)
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
	assetsdir := []string{"boat-sales", "boats", "boattrader"}
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
	btURL := "https://www.boattrader.com/boat/" + id
	btBoatPage, err := getPage(btBaseDir+"boattrader/"+id+".htm", btURL)
	if err != nil {
		return 0, err
	}
	btboatsJSON := btBoatPage.Find1ByRE(btBoatsJSONPattern, 1, "0", "0")
	boatType := boatTypePattern.FindStringSubmatch(btboatsJSON)[1]
	fieldXPath := func(name string) string {
		return `//div[@class='collapsible open']/table/tbody/tr/th[text()='` + name + `']/../td/text()`
	}
	boat := api.Boat{Sale: &api.BoatSale{}}
	var boatPage *page
	switch boatType {
	case "power":
		{
			url := btBaseURL + "power-boats/" + id
			boatPage, err = getPage(btBaseDir+"boats/"+id+".htm", url)
			if err != nil {
				return 0, err
			}
			boat = api.Boat{URLs: []string{url}}
			boat.Locomotion = "Power"
		}
	case "sail":
		{
			url := btBaseURL + "sailing-boats/" + id
			boatPage, err = getPage(btBaseDir+"boats/"+id+".htm", url)
			if err != nil {
				return 0, err
			}
			boat = api.Boat{URLs: []string{url}}
			boat.Locomotion = "Sail"
		}
	default:
		return 0, errors.New("boat not found")
	}
	boat.Year = boatPage.Int(boatPage.Find1(nil, fieldXPath("Year"), "", ""), nil)
	boat.Make = boatPage.Find1(nil, fieldXPath("Make"), "", "")
	boat.Model = boatPage.Find1(nil, fieldXPath("Model"), "", "")
	boat.Condition = boatPage.Find1(nil, fieldXPath("Condition"), "", "")
	boat.FuelType = boatPage.Find1(nil, fieldXPath("Fuel Type"), "", "")
	boat.Length = float32(boatPage.Float64(strings.Split(boatPage.Find1(nil, fieldXPath("Length"), "", ""), " ")[0], nil))
	boat.Beam = float32(boatPage.Float64(strings.Split(boatPage.Find1(nil, fieldXPath("Beam"), "", ""), " ")[0], nil))
	location := strings.Split(strings.Trim(boatPage.Find1(nil, fieldXPath("Location"), "", ""), ","), " ")
	boat.Location = &api.Contact{
		Type:  "Address",
		City:  location[0],
		State: location[1],
	}
	log.Println(boat)
	return 0, nil
}
