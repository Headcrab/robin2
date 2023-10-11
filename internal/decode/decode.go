package decode

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// LoadJSONData загружает JSON-данные из файла и заполняет поле TagClasses декодера.
//
// Функция принимает путь в качестве параметра, который указывает каталог, в котором находится JSON-файл.
// Возвращает ошибку в случае проблем с чтением файла или разбором его содержимого.
// Функция обновляет поле TagClasses декодера с помощью разобранных JSON-данных,
// а затем вызывает функцию prepareRegex для подготовки регулярных выражений, используемых декодером.
// Возвращает nil в случае успешного выполнения операции.
func (d *Decoder) LoadJSONData(path string) error {
	file, err := os.ReadFile(filepath.Join(path, "tag_classifier.json"))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	var tagClasses map[string]TagClass
	if err := json.Unmarshal(file, &tagClasses); err != nil {
		logger.Error(err.Error())
		return err
	}
	d.TagClasses = tagClasses
	d.prepareRegex()
	return nil
}

// DecodeTags декодирует теги в структуре Decoder.
//
// Функция DecodeTags инициализирует канал DecodedTagsChan и использует WaitGroup,
// чтобы параллельно декодировать каждый тег в срезе Tags. После завершения
// декодирования канал закрывается.
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

// decodeTag декодирует тег в структуре Decoder.
//
// Функция decodeTag принимает в качестве параметров тег и канал для передачи
// расшифрованной информации о теге. Она использует WaitGroup для синхронизации
// горутин и декодирует информацию о теге, сохраняя ее в карте decodedTag.
// Затем она отправляет расшифрованную информацию в переданный канал.
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

// decodeMatch декодирует значение на основе переданных групп и совпадений.
//
// Функция decodeMatch принимает в качестве параметров группу, значение
// TagClassOnMatch и срез совпадений. Она проверяет тип значения и возвращает
// соответствующее значение в зависимости от группы. Если значение является
// строкой и группа равна -1, то она возвращает это значение. Если значение
// является картой и группа не равна -1 и существует следующая группа в срезе
// совпадений, то она возвращает значение по этой группе. Если ни одно из
// условий не выполняется, она возвращает следующую группу совпадений.
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

// prepareRegex подготавливает регулярное выражение для декодирования.
//
// Функция prepareRegex принимает в качестве параметра строку с регулярным
// выражением и возвращает скомпилированное регулярное выражение и ошибку,
// если возникла. Она также устанавливает флаги компиляции для регулярного
// выражения, чтобы обеспечить его правильное декодирование.
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
