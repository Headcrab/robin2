package decode

import (
	"encoding/json"
	"os"
	"regexp"
	"robin2/internal/logger"
	"sync"
)

type TagClassOnMatch struct {
	Group int         `json:"group"`
	Value interface{} `json:"value"`
}

type TagClass struct {
	Regex     string                     `json:"regex"`
	OnMatch   map[string]TagClassOnMatch `json:"on_match"`
	RegexComp *regexp.Regexp
}

type Tag struct {
	Name string
}

type Decoder struct {
	TagClasses      map[string]TagClass
	Tags            []Tag
	DecodedTagsChan chan map[string]string
}

func (d *Decoder) LoadJSONData() {
	file, err := os.ReadFile("tag_classifier.json")
	if err != nil {
		logger.Fatal(err.Error())
	}

	var tagClasses map[string]TagClass
	if err := json.Unmarshal(file, &tagClasses); err != nil {
		logger.Fatal(err.Error())
	}
	d.TagClasses = tagClasses
	d.prepareRegex()
}

func (d *Decoder) DecodeTags() {
	d.DecodedTagsChan = make(chan map[string]string, len(d.Tags))
	var wg sync.WaitGroup
	for _, tag := range d.Tags {
		wg.Add(1)
		go d.decodeTag(tag, d.DecodedTagsChan, &wg)
	}
	wg.Wait()
	close(d.DecodedTagsChan)
}

func (d *Decoder) decodeTag(tag Tag, decodedTagsChan chan<- map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	decodedTag := make(map[string]string)
	decodedTag["tag_name"] = tag.Name

	for _, tagClass := range d.TagClasses {
		regexComp := tagClass.RegexComp
		onMatch := tagClass.OnMatch

		if match := regexComp.FindStringSubmatch(tag.Name); len(match) > 0 {
			for key, value := range onMatch {
				group := int(value.Group)
				decodedTag[key] = d.decodeMatch(group, value, match)
			}
		}
	}

	decodedTagsChan <- decodedTag
}

func (d *Decoder) decodeMatch(group int, value TagClassOnMatch, match []string) string {
	switch val := value.Value.(type) {
	case string:
		if group == -1 {
			return val
		}
	case map[string]interface{}:
		if group != -1 && group+1 < len(match) {
			if v, ok := val[match[group+1]].(string); ok {
				return v
			}
		}
	}
	if group != -1 && group+1 < len(match) {
		return match[group+1]
	}
	return ""
}

func (d *Decoder) prepareRegex() {
	var tagClasses map[string]TagClass = make(map[string]TagClass)
	for n, tagClass := range d.TagClasses {
		tagClasses[n] = TagClass{
			Regex:     tagClass.Regex,
			OnMatch:   tagClass.OnMatch,
			RegexComp: regexp.MustCompile(tagClass.Regex),
		}
	}
	d.TagClasses = tagClasses
}
