package sites

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"boatfuji.com/api"
)

var (
	btBaseURL     = "https://www.boats.com/"
	sourceBaseURL = "https://www.boattrader.com"
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
	assets := []string{"boats"}
	for _, assetdir := range assetsdir {
		os.MkdirAll(btBaseDir+assetdir, 0755)
	}
	switch url {
	case "":
		for _, asset := range assets {
			files, err := ioutil.ReadDir(btBaseDir + asset)
			if err != nil {
				return err
			}
			oldPercent := ""
			for fileIndex, file := range files {
				newPercent := strconv.Itoa(100 * fileIndex / len(files))
				if newPercent != oldPercent {
					log.Println(asset + " " + newPercent + "%")
				}
				oldPercent = newPercent
				var fileName = file.Name()
				if strings.HasSuffix(fileName, ".htm") {
					// id := strings.Split(filepath.Base(fileName), ".")[0]
					// if _, err = site.harvestX(asset, id); err != nil {
					// 	return err
					// }
				}
			}
		}
	case btBaseURL:
		for _, filter := range []string{"type-power", "type-sail", "type-small", "type-pwc"} {
			for page := 1; page < 99999; page++ {
				log.Println(fmt.Sprintf("%s/boat-sales/%d&%s.html", btBaseDir, page, filter))
				boatsPage, err := getPage(fmt.Sprintf("%s/boat-sales/%d&%s.html", btBaseDir, page, filter), fmt.Sprintf("%s/boats/%s/page-%d/", sourceBaseURL, filter, page))
				if err != nil {
					return err
				}
				bsBoatIDs := boatsPage.FindN(nil, "//li/@data-listing-id", 0, 99999999, "", "")
				log.Println(bsBoatIDs)
			}
		}
	default:
		return nil
	}
	return nil
}
