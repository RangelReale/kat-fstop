package kat

import (
	"errors"
	"fmt"
	gq "github.com/PuerkitoBio/goquery"
	"github.com/RangelReale/filesharetop/lib"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type KATSort string

const (
	KATSORT_SEEDERS  KATSort = "seeders"
	KATSORT_LEECHERS KATSort = "leechers"
)

type KATSortBy string

const (
	KATSORTBY_ASCENDING  KATSortBy = "asc"
	KATSORTBY_DESCENDING KATSortBy = "desc"
)

type KATParser struct {
	List   map[string]*fstoplib.Item
	config *Config
	logger *log.Logger
}

func NewKATParser(config *Config, l *log.Logger) *KATParser {
	return &KATParser{
		List:   make(map[string]*fstoplib.Item),
		config: config,
		logger: l,
	}
}

func (p *KATParser) CategoryMap() (*fstoplib.CategoryMap, error) {
	return &fstoplib.CategoryMap{
		"TV":     []string{"tv"},
		"GAMES":  []string{"games"},
		"APPS":   []string{"applications"},
		"OTHER":  []string{"other"},
		"MOVIES": []string{"movies"},
		"MUSIC":  []string{"music"},
		"BOOKS":  []string{"books"},
		"ANIME":  []string{"anime"},
		"XXX":    []string{"xxx"},
	}, nil
}

func (p *KATParser) Parse(sort KATSort, sortby KATSortBy, pages int) error {

	if pages < 1 {
		return errors.New("Pages must be at least 1")
	}

	catmap, err := p.CategoryMap()
	if err != nil {
		return err
	}

	for _, catdata := range *catmap {

		posct := int32(0)
		for pg := 1; pg <= pages; pg++ {
			var doc *gq.Document
			var e error

			// download the page
			var u *url.URL
			if pg > 1 {
				u, e = url.Parse(fmt.Sprintf("https://kat.cr/%s/%d/?field=%s&sorder=%s", catdata[0], pg, sort, sortby, pg))
			} else {
				u, e = url.Parse(fmt.Sprintf("https://kat.cr/%s/?field=%s&sorder=%s", catdata[0], sort, sortby, pg))
			}
			if e != nil {
				return e
			}

			cookies, _ := cookiejar.New(nil)
			/*
				cookies.SetCookies(u, []*http.Cookie{
				})
			*/

			client := &http.Client{
				Jar: cookies,
			}

			req, e := http.NewRequest("GET", u.String(), nil)
			if e != nil {
				return e
			}
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.3; WOW64; Trident/7.0; MDDCJS; rv:11.0) like Gecko")

			resp, e := client.Do(req)
			if e != nil {
				return e
			}

			// parse the page
			if doc, e = gq.NewDocumentFromResponse(resp); e != nil {
				return e
			}

			/*
				// regular expressions
				re_id := regexp.MustCompile("/torrent/(\\d+)-")
				re_category := regexp.MustCompile("fa-(\\w+)")
				re_category_cat3 := regexp.MustCompile("text-pink")
			*/
			re_adddate := regexp.MustCompile("(\\d+)\\W+(\\w+)")

			// Iterate on each record
			doc.Find("#mainSearchTable table.data tr[id^=torrent]").Each(func(i int, s *gq.Selection) {
				//var se error

				link := s.Find("td > div.torrentname div.torType a.cellMainLink").First()
				if link.Length() == 0 {
					//p.logger.Println("ERROR: Link not found")
					return
				}

				href, hvalid := link.Attr("href")
				if !hvalid || href == "" {
					p.logger.Println("ERROR: Link not found")
					return
				}

				hu, se := url.Parse(href)
				if se != nil {
					p.logger.Printf("ERROR: %s", se)
					return
				}
				hu.Scheme = "https"
				hu.Host = "kat.cr"

				//fmt.Printf("Link: %s\n", hu.String())

				lid, lvalid := s.Attr("id")
				if !lvalid || lid == "" {
					p.logger.Println("ERROR: ID not found")
					return
				}

				//fmt.Printf("ID: %s\n", lid)

				seeder := s.Find("td").Eq(-2)
				if seeder.Length() == 0 {
					p.logger.Println("ERROR: Seeder not found")
					return
				}

				leecher := s.Find("td").Eq(-1)
				if leecher.Length() == 0 {
					p.logger.Println("ERROR: Leecher not found")
					return
				}
				/*
					complete := s.Find("td").Eq(-1)
					if complete.Length() == 0 {
						p.logger.Println("ERROR: Complete not found")
						return
					}
				*/
				comments := s.Find("td .iaconbox .iconvalue")
				/*
					if comments.Length() == 0 {
						p.logger.Println("ERROR: Comments not found")
						return
					}
				*/

				adddate := s.Find("td").Eq(-3)
				if adddate.Length() == 0 {
					p.logger.Println("ERROR: Adddate not found")
					return
				}

				nseeder, se := strconv.ParseInt(seeder.Text(), 10, 32)
				if se != nil {
					p.logger.Printf("ERROR: %s", se)
					return
				}

				nleecher, se := strconv.ParseInt(leecher.Text(), 10, 32)
				if se != nil {
					p.logger.Printf("ERROR: %s", se)
					return
				}

				/*
					ncomplete := int64(0)
					if complete.Length() > 0 {
						ncomplete, se = strconv.ParseInt(complete.Text(), 10, 32)
						if se != nil {
							p.logger.Printf("ERROR: %s", se)
							return
						}
					}
				*/
				ncomments := int64(0)
				if comments.Length() > 0 && comments.Text() != "---" {
					ncomments, se = strconv.ParseInt(comments.Text(), 10, 32)
					if se != nil {
						p.logger.Printf("ERROR: %s", se)
						return
					}
				}

				//fmt.Printf("@@ %d - %d - %d - %d\n", nseeder, nleecher, ncomplete, ncomments)

				adddatematch := re_adddate.FindAllStringSubmatch(strings.TrimSpace(adddate.Text()), -1)
				if adddatematch == nil || len(adddatematch) < 1 || len(adddatematch[0]) < 3 {
					p.logger.Printf("Adddate not match")
					return
				}

				nduration, se := strconv.ParseInt(adddatematch[0][1], 10, 32)
				if se != nil {
					p.logger.Printf("ERROR parsing duration: %s", se)
					return
				}

				nadddate := time.Now()
				switch adddatematch[0][2] {
				case "min", "minute", "minutes":
					nadddate = nadddate.Add(-1 * time.Minute * time.Duration(nduration))
				case "hour", "hours":
					nadddate = nadddate.Add(-1 * time.Hour * time.Duration(nduration))
				case "day", "days":
					nadddate = nadddate.Add(-1 * 24 * time.Hour * time.Duration(nduration))
				case "week", "weeks":
					nadddate = nadddate.Add(-1 * 7 * 24 * time.Hour * time.Duration(nduration))
				case "month", "months":
					nadddate = nadddate.Add(-1 * 30 * 24 * time.Hour * time.Duration(nduration))
				case "year", "years":
					nadddate = nadddate.Add(-1 * 365 * 24 * time.Hour * time.Duration(nduration))
				default:
					p.logger.Printf("ERROR determining duration: %s", adddatematch[0][2])
					return
				}

				//fmt.Println(nadddate.String())

				//fmt.Printf("%s: %s\n", strings.TrimSpace(link.Text()), lid)
				item, ok := p.List[lid]
				if !ok {
					item = fstoplib.NewItem()
					item.Id = lid
					item.Title = strings.TrimSpace(link.Text())
					item.Link = hu.String()
					item.Count = 0
					item.Category = catdata[0]
					item.AddDate = nadddate.Format("2006-01-02")
					item.Seeders = int32(nseeder)
					item.Leechers = int32(nleecher)
					//item.Complete = int32(ncomplete)
					item.Comments = int32(ncomments)
					p.List[lid] = item
				}
				item.Count++
				posct++
				if sort == KATSORT_SEEDERS {
					item.SeedersPos = posct
				} else if sort == KATSORT_LEECHERS {
					item.LeechersPos = posct
					//} else if sort == KATSORT_COMPLETE {
					//item.CompletePos = posct
				}
			})
		}
	}

	return nil
}
