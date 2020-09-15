package sites

import (
	"os"

	"boatfuji.com/api"
)

var btBaseURL = "https://www.boats.com/"
var btBaseDir = "harvest/www.boats.com/"

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
	os.MkdirAll(btBaseDir+"boat-rentals", 0755)
	os.MkdirAll(btBaseDir+"boats", 0755)
	os.MkdirAll(btBaseDir+"users", 0755)
	return nil
}
