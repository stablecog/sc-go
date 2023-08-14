package utils

import (
	"net"
	"path"

	"github.com/oschwald/geoip2-golang"
	"github.com/stablecog/sc-go/log"
)

type GeoIP struct {
	db *geoip2.Reader
}

func NewGeoIPService(mock bool) (*GeoIP, error) {
	dbName := "GeoLite2-Country.mmdb"
	if mock {
		dbName = "GeoLite2-City-Test.mmdb"
	}
	db, err := geoip2.Open(path.Join(RootDir(), "utils", dbName))
	if err != nil {
		return nil, err
	}
	return &GeoIP{db: db}, nil
}

func (g *GeoIP) Close() {
	g.db.Close()
}

func (g *GeoIP) GetCountryFromIP(ipAddr string) (country string, err error) {
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(ipAddr)
	record, err := g.db.Country(ip)
	if err != nil {
		log.Error("Error getting country from geoip", "err", err)
		return "", err
	}
	return record.Country.IsoCode, nil
}
