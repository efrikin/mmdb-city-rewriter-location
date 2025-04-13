package config

type Config struct {
	MMDBFile         string `env:"MMDB_FILE" envDefault:"Custom-GeoIP2-City.mmdb"`
	MMDBPostfix      string `env:"MMDB_POSTFIX" envDefault:"fix"`
	NumNetwork       int
}
