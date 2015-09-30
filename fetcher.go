package kat

import (
	"github.com/RangelReale/filesharetop/lib"
	"io/ioutil"
	"log"
)

type Fetcher struct {
	logger *log.Logger
	config *Config
}

func NewFetcher(config *Config) *Fetcher {
	return &Fetcher{
		logger: log.New(ioutil.Discard, "", 0),
		config: config,
	}
}

func (f *Fetcher) ID() string {
	return "KAT"
}

func (f *Fetcher) SetLogger(l *log.Logger) {
	f.logger = l
}

func (f *Fetcher) Fetch() (map[string]*fstoplib.Item, error) {
	parser := NewKATParser(f.config, f.logger)

	// parse 4 pages ordered by seeders
	err := parser.Parse(KATSORT_SEEDERS, KATSORTBY_DESCENDING, 4)
	if err != nil {
		return nil, err
	}

	// parse 2 pages ordered by leechers
	err = parser.Parse(KATSORT_LEECHERS, KATSORTBY_DESCENDING, 2)
	if err != nil {
		return nil, err
	}

	return parser.List, nil
}

func (f *Fetcher) CategoryMap() (*fstoplib.CategoryMap, error) {
	parser := NewKATParser(f.config, f.logger)
	return parser.CategoryMap()
}
