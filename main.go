package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/efrikin/mmdb-city-rewriter-location/internal/config"
)

func main() {
	var cfg config.Config

	// MMDB can be defined eiher via MMDB_FILE environment variable or -f flag
	// flag have priority
	flag.StringVar(&cfg.MMDBPostfix, "p", "fix", "DB postfix")
	flag.StringVar(&cfg.MMDBFile, "f", "Custom-GeoIP2-City.mmdb", "The path to the MMDB file. Required.")

	err := env.Parse(&cfg)
	flag.Parse()

	if err != nil {
		log.Panic(err)
	}

	log.Printf("DB %s will be used", cfg.MMDBFile)

	// mmdbwriter library doesn't have getting all networks
	// https://github.com/maxmind/mmdbwriter/blob/main/tree.go#L183
	mmdb, err := maxminddb.Open(cfg.MMDBFile)
	if err != nil {
		log.Fatal(err)
	}

	// https://go.dev/ref/spec#Defer_statements
	defer func() {
		if err := mmdb.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()

	writer, err := mmdbwriter.Load(cfg.MMDBFile, mmdbwriter.Options{
		IncludeReservedNetworks: true,
	},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get all netwoks and change data type for latitude/longitude
	for network := range mmdb.Networks() {
		ip, net, _ := net.ParseCIDR(network.Prefix().String())
		_, data := writer.Get(ip)

		if _, locationIsExist := data.(mmdbtype.Map)["location"]; locationIsExist {
			if latitude, latitudeIsExist := data.(mmdbtype.Map)["location"].(mmdbtype.Map)["latitude"]; latitudeIsExist {
				if longitude, longitudeIsExist := data.(mmdbtype.Map)["location"].(mmdbtype.Map)["longitude"]; longitudeIsExist {
					location := mmdbtype.Map{
						"location": mmdbtype.Map{
							"latitude":  mmdbtype.Float64(latitude.(mmdbtype.Float32)),
							"longitude": mmdbtype.Float64(longitude.(mmdbtype.Float32)),
						},
					}
					if err := writer.InsertFunc(net, inserter.DeepMergeWith(location)); err != nil {
						log.Fatal(err)
					}
				}
			} else {
				log.Printf("Latitude not found for net: %s", net)
			}
		}
		cfg.NumNetwork++
	}

	log.Printf("%d networks was processed", cfg.NumNetwork)

	fd, err := os.Create(fmt.Sprintf("%s.%s", cfg.MMDBFile, cfg.MMDBPostfix))
	if err != nil {
		log.Fatal(err)
	}

	_, err = writer.WriteTo(fd)
	if err != nil {
		log.Fatal(err)
	}
}
