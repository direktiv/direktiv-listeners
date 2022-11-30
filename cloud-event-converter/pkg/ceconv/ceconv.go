package ceconv

import (
	"bytes"
	"encoding/json"

	"github.com/itchyny/gojq"
)

type Condition struct {
	src   string
	query *gojq.Query
}

type Modifier struct {
	c *Condition
}

func LoadCondition(s string) (*Condition, error) {

	x, err := gojq.Parse(s)
	if err != nil {
		return nil, err
	}

	return &Condition{
		src:   s,
		query: x,
	}, nil
}

func MapFromByteSlice(b []byte) (map[string]interface{}, error) {

	m := make(map[string]interface{})
	err := json.NewDecoder(bytes.NewReader(b)).Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Condition) Evaluate(m map[string]interface{}) (bool, error) {

	iter := c.query.Run(m)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return false, err
		}

		if b, ok := v.(bool); ok {
			if b {
				return true, nil
			}
		}
	}

	return false, nil
}

func LoadModifier(s string) (*Modifier, error) {

	var err error

	m := new(Modifier)
	m.c, err = LoadCondition(s)
	if err != nil {
		return nil, err
	}

	return m, nil

}

func (m *Modifier) Modify(x map[string]interface{}) ([]string, error) {

	out := make([]string, 0)
	iter := m.c.query.Run(x)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return nil, err
		}

		b, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return nil, err
		}

		out = append(out, string(b))
	}

	return out, nil
}
