package sites

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"boatfuji.com/api"
)

var (
	btBaseURL     = "https://www.boats.com/"
	sourceBaseURL = "https://www.boattrader.com"
	btURLPattern  = regexp.MustCompile(`^https://www\.boats\.com/boat/([0-9]*)`)
	btBaseDir     = "/Users/ankitagarwal/ankit_code/boat-golang/harvest/www.boats.com/"
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
	assetsdir := []string{"boat-sales", "boats"}
	for _, assetdir := range assetsdir {
		os.MkdirAll(btBaseDir+assetdir, 0755)
	}
	switch url {
	case btBaseURL:
		boatsByFilters = map[string][]string{}
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
	return nil
}

// func (site *Boats) harvestX(x, id string) (int64, error) {
// 	return site.harvestBoat(id, 0, nil)
// }

// func (site *Boats) harvestBoat(x, id string) (int64, error) {
// 	return site.harvestBoat(id, 0, nil)
// }
