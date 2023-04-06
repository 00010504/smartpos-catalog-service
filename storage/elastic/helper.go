package elastic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/pkg/errors"
)

func exists(es *elasticsearch.Client, indexName string) bool {
	res, err := es.Indices.Get([]string{indexName})

	return err == nil && res != nil && res.StatusCode != 404
}

func checkResponseCodeToSuccess(code int) error {

	if !(code >= 200 && code < 300) {
		return errors.New("Response status is not Success")
	}

	return nil
}

func splitSearchInput(input string) []H {

	input = strings.ToLower(input)

	res := make([]H, 0)

	if len(input) == 0 {
		return res
	}

	inputList := strings.Split(input, " ")

	for _, v := range inputList {
		h := H{
			"query_string": H{
				"fields":           []string{"name.keyword", "sku.keyword^4", "description.keyword", "barcode.keyword"},
				"query":            fmt.Sprintf("*%s*", v),
				"default_operator": "AND",
			},
		}

		res = append(res, h)

	}

	return res

}

func getSearchString(input string) string {
	specialCharacteristics := []string{"-", "+", "=", "&&", "||", ">", "<", "!", "(", ")", "{", "}", "[", "]", "^", `"`, "~", "*", "?", ":", `\`, "/"}

	for _, ch := range specialCharacteristics {
		input = strings.ReplaceAll(input, ch, " ")
	}

	re := regexp.MustCompile(`\s+`)
	input = re.ReplaceAllString(input, " ")
	input = strings.TrimSpace(input)

	strArr := strings.Split(input, " ")

	for i := 0; i < len(strArr); i++ {
		strArr[i] = fmt.Sprintf("*%s*", strArr[i])
	}

	result := strings.Join(strArr, " ")

	return result
}
