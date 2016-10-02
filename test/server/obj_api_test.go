package server

import (
	"testing"
	//neturl "net/url"
	"fmt"

	"github.com/go-resty/resty"
	"github.com/stretchr/testify/assert"
	"github.com/toolkits/file"
	"github.com/chaosxu/nerv/lib/model"
	"reflect"
	"encoding/json"
	"strings"
	"time"
)

func TestClassRest(t *testing.T) {
	testCRUD(t, "Class", "classes/host/host.json")
}

func TestUpdateAddChild(t *testing.T) {
	data := create(t, "Class", "classes/host/host.json").(*model.Class)
	data.Operations = append(data.Operations, model.Operation{Name:"updateAddChile", Type:"go", Implementor:"test"})
	update(t, "Class", data)
}

func TestUpdateRemoveChild(t *testing.T) {
	data := create(t, "Class", "classes/host/host.json").(*model.Class)
	data.Operations = append(data.Operations, model.Operation{Name:"updateAddChile", Type:"go", Implementor:"test"})
	data = update(t, "Class", data).(*model.Class)

	fmt.Printf("%d\n", len(data.Operations))
	data.Operations = data.Operations[:len(data.Operations) - 1]
	fmt.Printf("%d\n", len(data.Operations))
	update(t, "Class", data)
}

func TestUpdateAddRemoveAddChild(t *testing.T) {
	data := create(t, "Class", "classes/host/host.json").(*model.Class)
	data.Operations = append(data.Operations, model.Operation{Name:"uar-a", Type:"go", Implementor:"test"})

	time.Sleep(time.Second)
	data = update(t, "Class", data).(*model.Class)

	fmt.Printf("%d\n", len(data.Operations))
	data.Operations = data.Operations[:len(data.Operations) - 1]
	fmt.Printf("%d\n", len(data.Operations))
	data.Operations = append(data.Operations, model.Operation{Name:"uar-ara", Type:"go", Implementor:"test"})

	time.Sleep(time.Second)
	update(t, "Class", data)
}

func TestUpdateUpdateChild(t *testing.T) {

}

func testCRUD(t *testing.T, class string, dataPath string) {
	data := create(t, class, dataPath)

	find(t, class)

	v := reflect.ValueOf(data).Elem()
	id := v.FieldByName("ID").Interface()
	data = getAndPreLoad(t, class, id)

	time := v.FieldByName("UpdatedAt").Interface().(time.Time)
	fmt.Println(time)
	data = update(t, class, data)

	remove(t, class, id)

	getNil(t, class, id)
}

func find(t *testing.T, class string) {
	logCodeLine()
	md := model.Models[class]

	url := fmt.Sprintf("http://localhost:3333/objs/%s?where=name=?&values=/host/Host", class)
	fmt.Println(url)
	res, err := resty.R().
			SetHeader("Content-Type", "application/json").
			Get(url)
	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
	b := res.Body();
	//fmt.Printf("body\n %s\n", string(b))

	data := md.NewSlice()
	err = json.Unmarshal(b, data)
	assert.Nil(t, err)

	fmt.Println("find")
	fmt.Println(data)
}

func remove(t *testing.T, class string, id interface{}) {
	logCodeLine()
	var (
		err error
		res *resty.Response
	)

	res, err = resty.R().
			SetHeader("Content-Type", "application/json").
			Delete(fmt.Sprintf("http://localhost:3333/objs/%s/%d", class, id))

	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
}

func update(t *testing.T, class string, data interface{}) interface{} {
	logCodeLine()
	var (
		body []byte
		err error
		res *resty.Response
	)

	if body, err = json.Marshal(data); err != nil {
		assert.NotNil(t, err, err.Error())
	}
	res, err = resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Put("http://localhost:3333/objs/" + class)

	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
	b := res.Body();
	//fmt.Printf("body\n %s\n", string(b))
	md := model.Models[class]
	updated := md.New()
	err = json.Unmarshal(b, updated)
	assert.Nil(t, err)
	return data
}

func create(t *testing.T, class string, dataPath string) interface{} {
	logCodeLine()
	var (
		body string
		err error
		res *resty.Response
	)
	body, err = file.ToTrimString(dataPath)
	assert.Nil(t, err)
	res, err = resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post("http://localhost:3333/objs/" + class)

	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
	b := res.Body();
	//fmt.Printf("body\n %s\n", string(b))
	md := model.Models[class]
	data := md.New()
	err = json.Unmarshal(b, data)
	assert.Nil(t, err)
	return data
}

func getAndPreLoad(t *testing.T, class string, id interface{}) interface{} {
	logCodeLine()
	md := model.Models[class]
	assList := associations(reflect.TypeOf(md.Type).Elem(), "", []string{})
	ass := strings.Join(assList, ",")
	var url string
	if ass == "" {
		url = fmt.Sprintf("http://localhost:3333/objs/%s", class, id)
	} else {
		url = fmt.Sprintf("http://localhost:3333/objs/%s/%d?associations=%s", class, id, ass)
	}
	res, err := resty.R().
			SetHeader("Content-Type", "application/json").
			Get(url)
	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
	b := res.Body();
	//fmt.Printf("body\n %s\n", string(b))

	data := md.New()
	err = json.Unmarshal(b, data)
	assert.Nil(t, err)

	v := reflect.ValueOf(data).Elem()
	for _, as := range assList {
		assertAssociations(t, v, as)
	}
	return data
}

func getNil(t *testing.T, class string, id interface{}) interface{} {
	logCodeLine()
	md := model.Models[class]
	assList := associations(reflect.TypeOf(md.Type).Elem(), "", []string{})
	ass := strings.Join(assList, ",")
	var url string
	if ass == "" {
		url = fmt.Sprintf("http://localhost:3333/objs/%s", class, id)
	} else {
		url = fmt.Sprintf("http://localhost:3333/objs/%s/%d?associations=%s", class, id, ass)
	}
	res, err := resty.R().
			SetHeader("Content-Type", "application/json").
			Get(url)
	if err != nil {
		assert.Nil(t, err, err.Error())
	}
	assert.Equal(t, 200, res.StatusCode())
	b := res.Body();
	//fmt.Printf("body\n %s\n", string(b))

	data := md.New()
	err = json.Unmarshal(b, data)
	assert.Nil(t, err)
	assert.Equal(t, "null\n", string(b))
	fmt.Println("return body:" + string(b))

	return data
}

func assertAssociations(t *testing.T, v reflect.Value, ass string) {
	assList := strings.Split(ass, ".")
	as := v.Type().Name()
	cv := v
	for _, ass := range assList {
		switch  cv.Kind(){
		case reflect.Slice:
			cv = cv.FieldByName(ass)
			as = as + "." + ass
			fmt.Printf("assert %s\n", as)
			if cv.Len() > 0 {
				cv = cv.Index(0)
				assert.NotNil(t, cv.Interface(), as + " is nil")
			}
		case reflect.Struct:
			cv = cv.FieldByName(ass)
			as = as + "." + ass
			fmt.Printf("assert %s\n", as)
			assert.NotNil(t, cv.Interface(), as + " is nil")
		}
	}
}

func associations(t reflect.Type, parent string, ass []string) []string {
	n := t.NumField()
	if parent != "" {
		parent = parent + "."
	}
	for i := 0; i < n; i++ {
		f := t.Field(i)

		switch f.Type.Kind() {
		case reflect.Slice:
			fmt.Printf("Name:%s,Type:%s,Kind:%s\n", f.Name, f.Type.Elem().Name(), f.Type.Kind())
			ass = append(ass, f.Name)
			return associations(f.Type.Elem(), parent + f.Name, ass)
		case reflect.Struct:
			fmt.Printf("Name:%s,Type:%s,Kind:%s\n", f.Name, f.Type.Name(), f.Type.Kind())
			//Not embedded type
			if f.Name != f.Type.Name() {
				ass = append(ass, f.Name)
				return associations(f.Type, parent + f.Name, ass)
			}
		}
	}

	return ass
}

