package indicphone

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/knadh/dictpress/internal/data"
	"github.com/knadh/knphone"
	"gitlab.com/joice/mlphone-go"
	"github.com/soumendrak/odiphone" 
)

// IndicPhone is a phonetic tokenizer that generates phonetic tokens for
// Indian languages. It is similar to Metaphone for English.
type IndicPhone struct {
	kn *knphone.KNphone
	ml *mlphone.MLPhone
	od *odiphone.ODIphone
}

// New returns a new instance of the Kannada tokenizer.
func New() *IndicPhone {
	return &IndicPhone{
		kn: knphone.New(),
		ml: mlphone.New(),
		od: odiphone.New(),
	}
}

// ToTokens tokenizes a string and a language returns an array of tsvector tokens.
// eg: [KRM0 KRM] or [KRM:2 KRM:1] with weights.
func (ip *IndicPhone) ToTokens(s string, lang string) ([]string, error) {
	if lang != "kannada" && lang != "malayalam"  && lang != "odia" {
		return nil, errors.New("unknown language to tokenize")
	}

	var (
		chunks = strings.Split(s, " ")
		tokens = make([]data.Token, 0, len(chunks)*3)

		key0, key1, key2 string
	)
	for _, c := range chunks {
		switch lang {
		case "kannada":
			key0, key1, key2 = ip.kn.Encode(c)
		case "malayalam":
			key0, key1, key2 = ip.ml.Encode(c)
		case "odia":
			key0, key1, key2 = ip.od.Encode(c)
		}

		if key0 == "" {
			continue
		}

		tokens = append(tokens,
			data.Token{Token: key0, Weight: 3},
			data.Token{Token: key1, Weight: 2},
			data.Token{Token: key2, Weight: 1})
	}

	return data.TokensToTSVector(tokens), nil
}

// ToQuery tokenizes a Kannada string into Romanized (knphone) Postgres
// tsquery string.
func (ip *IndicPhone) ToQuery(s string, lang string) (string, error) {
	var key0, key1, key2 string

	switch lang {
	case "kannada":
		key0, key1, key2 = ip.kn.Encode(s)
	case "malayalam":
		key0, key1, key2 = ip.ml.Encode(s)
	case "odia":
		key0, key1, key2 = ip.od.Encode(s)
	}

	if key0 == "" {
		return "", nil
	}

	tokens := slices.Compact([]string{key2, key1, key0})

	if len(tokens) == 3 {
		return fmt.Sprintf("%s | %s", key2, key1), nil
	}

	return tokens[0], nil
}
