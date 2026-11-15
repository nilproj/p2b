package generator

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

const filename = "p2b_gen.go"

var (
	//go:embed p2b_gen.gotpl
	p2bText string

	p2bTpl = template.Must(template.New("p2b").Parse(p2bText))
)

var (
	modelFields = map[string]map[string]Field{}
	relations   = map[string]Field{}
)

func normalize(datamodel Datamodel) {
	for _, model := range datamodel.Models {
		if modelFields[model.Name] == nil {
			modelFields[model.Name] = map[string]Field{}
		}
		for i, f := range model.Fields {
			model.Fields[i].model = &model

			modelFields[model.Name][f.Name] = f
			if n := f.RelationName; n != "" {
				if len(f.RelationFromFields) > 0 {
					relations[n] = f
				}
			}
		}
	}
}

func New(params []byte) (*Generator, error) {
	var g Generator
	if err := json.Unmarshal(params, &g); err != nil {
		return nil, err
	}

	normalize(g.DMMF.Datamodel)

	return &g, nil

}

type Generator struct {
	Config GeneratorConfig `json:"generator"`
	DMMF   DMMF            `json:"dmmf"`
}

func (g *Generator) Generate() error {
	dir := g.Config.Output.Value
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	output := filepath.Join(dir, filename)
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	if err := p2bTpl.Execute(&buf, g.DMMF.Datamodel); err != nil {
		return err
	}

	b, err := imports.Process(output, buf.Bytes(), nil)
	if err != nil {
		return err
	}

	_, err = file.Write(b)
	return err
}

type GeneratorConfig struct {
	Output EnvValue `json:"output"`
}

type EnvValue struct {
	Value string `json:"value"`
}

type DMMF struct {
	Datamodel Datamodel `json:"datamodel"`
}

type Datamodel struct {
	Enums  []Enum  `json:"enums"`
	Models []Model `json:"models"`
}

type Enum struct {
	IName
	Values []IName `json:"values"`
}

type Model struct {
	IName
	Fields     []Field `json:"fields"`
	PrimaryKey Key     `json:"primaryKey"`
}

type Key struct {
	Fields []string `json:"fields"`
}

type Field struct {
	IName
	Type               FieldType `json:"type"`
	List               bool      `json:"isList"`
	ID                 bool      `json:"isID"`
	Required           bool      `json:"isRequired"`
	Default            any       `json:"default"`
	RelationName       string    `json:"relationName"`
	RelationFromFields []string  `json:"relationFromFields"`
	RelationToFields   []string  `json:"relationToFields"`
	Documentation      string    `json:"documentation"`

	model *Model
}

func (f Field) GoType() string {
	t := f.Type
	if f.List {
		return fmt.Sprintf("[]%s", t)
	} else if !f.Required || f.RelationName != "" {
		return fmt.Sprintf("*%s", t)
	}
	return t.String()
}

func (f Field) Tags() string {
	tags := []string{}

	if f.RelationName != "" {
		if len(f.RelationFromFields) > 0 {
			tags = append(tags, "rel:belongs-to")
			for i, ff := range f.RelationFromFields {
				fk := modelFields[f.model.Name][ff].DBName
				pk := modelFields[string(f.Type)][f.RelationToFields[i]].DBName
				tags = append(tags, fmt.Sprintf("join:%s=%s", fk, pk))
			}
		} else {
			tags = append(tags, "rel:has-many")
			for i, ff := range relations[f.RelationName].RelationToFields {
				fk := modelFields[f.model.Name][ff].DBName
				pk := modelFields[string(f.Type)][relations[f.RelationName].RelationFromFields[i]].DBName
				tags = append(tags, fmt.Sprintf("join:%s=%s", fk, pk))
			}
		}
	} else {
		tags = append(tags, f.DBName)
		if f.ID || slices.Contains(f.model.PrimaryKey.Fields, f.Name) {
			tags = append(tags, "pk")
		}
		switch v := f.Default.(type) {
		case nil:
		case map[string]any:
			if v["name"] == "autoincrement" {
				tags = append(tags, "autoincrement")
			}
		default:
			tags = append(tags, "nullzero")
		}
	}

	return strings.Join(tags, ",")
}

type IName struct {
	Name   string `json:"name"`
	DBName string `json:"dbName"`
}

type FieldType string

func (t FieldType) String() string {
	s := string(t)
	if fs, ok := p2bTypes[s]; ok {
		return fs
	}
	return s
}

var p2bTypes = map[string]string{
	"Int":      "int32",
	"String":   "string",
	"Decimal":  "float64",
	"DateTime": "time.Time",
}

type FieldValue struct {
	Name string `json:"name"`
}
