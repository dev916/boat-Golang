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
	url := "https://www.boattrader.com/boat/" + id
	boat := api.Boat{URLs: []string{url}, Sale: &api.BoatSale{}}
	btBoatPage, err := getPage(btBaseDir+"boattrader/"+id+".htm", url)
	if err != nil {
		return 0, err
	}
	btboatsJSON := btBoatPage.Find1ByRE(btBoatsJSONPattern, 1, "0", "0")
	boatType := boatTypePattern.FindStringSubmatch(btboatsJSON)[1]
	switch boatType {
	case "power":
		{
			fieldXPath := func(name string) string {
				return `//div[@class='collapsible open']/table/tbody/tr/th[text()=` + name + `]/td/child::text()`
			}

			url := "https://www.boats.com/power-boats/" + id
			boatPage, err := getPage(btBaseDir+"boats/"+id+".htm", url)
			if err != nil {
				return 0, err
			}
			boat.Year = boatPage.Int(boatPage.Find0or1(nil, fieldXPath("Model"), "", ""), nil)
			log.Println(boat.Year)
		}
	case "sail":
		{
			url := "https://www.boats.com/sailing-boats/" + id
			_, err := getPage(btBaseDir+"boats/"+id+".htm", url)
			if err != nil {
				return 0, err
			}
		}
	}
	return 0, nil
}
