// models.board.go

package model

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type RenderItem struct {
	Name        string
	Value       string
	Default     string
	Description string
	Title       string
	FormName    string
}

// Convert the object to a renderable payload
func getRenderItem(r interface{}) ([]RenderItem, error) {

	val := reflect.ValueOf(r).Elem()

	m := []RenderItem{}
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if typeField.Name == "Model" {
			continue
		}
		tag := typeField.Tag
		if _, ok := tag.Lookup("table_default"); !ok {
			return nil, fmt.Errorf("Field %s is missing tag table_default", typeField.Name)
		}
		if _, ok := tag.Lookup("table_descr"); !ok {
			return nil, fmt.Errorf("Field %s is missing tag table_descr", typeField.Name)
		}

		ItemValue, ok := valueField.Interface().(string)
		if !ok {
			ItemValue = ""
		}
		m = append(m, RenderItem{
			Name:        typeField.Name,
			Value:       ItemValue,
			Default:     tag.Get("table_default"),
			Description: tag.Get("table_descr"),
			Title:       tag.Get("table_title"),
			FormName:    tag.Get("form"),
		})
	}

	return m, nil
}

type RenderList struct {
	ID    uint
	Value []string
}

// Convert the objects to a renderable payload
func getRenderList(l interface{}) (RenderList, []RenderList, error) {

	h := RenderList{}
	m := []RenderList{}
	arg := reflect.ValueOf(l)
	if arg.Kind() != reflect.Slice {
		return h, nil, fmt.Errorf("Argument is not a slice")
	}

	if arg.Len() > 0 {
		val := arg.Index(0)

		for i := 0; i < val.NumField(); i++ {
			typeField := val.Type().Field(i)

			t, ok := typeField.Tag.Lookup("table_list")
			if !ok {
				continue
			}

			h.Value = append(h.Value, t)
		}
	}
	for j := 0; j < arg.Len(); j++ {
		val := arg.Index(j)

		renderList := RenderList{}
		for i := 0; i < val.NumField(); i++ {
			valueField := val.Field(i)
			typeField := val.Type().Field(i)
			if typeField.Name == "Model" {
				model := valueField.Interface().(gorm.Model)
				renderList.ID = model.ID
				continue
			}
			tag := typeField.Tag
			if _, ok := tag.Lookup("table_list"); !ok {
				continue
			}
			ItemValue, ok := valueField.Interface().(string)
			if !ok {
				ItemValue = ""
			}

			renderList.Value = append(renderList.Value, ItemValue)
		}
		if len(renderList.Value) > 0 {
			m = append(m, renderList)
		}
	}
	return h, m, nil
}
