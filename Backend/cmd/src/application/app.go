package application

import "github.com/crcaniullan-commits/Tally/cmd/src/routes"

type application struct {
	config config
}

type config struct {
	addr   string
	apiURL string
	routes routes.Routes
}

func main() {

}
