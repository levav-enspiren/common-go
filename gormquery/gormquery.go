package gormquery

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gitea.greatics.net/common-go/skmap"
	"gorm.io/gorm"
)

type QueryFactory struct {
	Query *gorm.DB
}

/*
Analyze field array and apply to query

Input

	fields: field array
	compulsoryFields: fields that must be applied when any field is applied
	relations: relational field and its depending fields

Field array format

	"fieldName": tipical field name
	"relation.subField": declare sub-field selection of relational field
	"relation": relational field name with no sub-field selection (all fields)

Output

	relationFieldMap
		key (string): relation field name
		value ([]string): field array to be applied on relation query

Example

	// Create query factory
	qf := gormquery.QueryFactory{ Query: models.DbServer.Model(&model{}) }
	// To select normal field "name" and relational field "item",
	// with item field selection "item_name"
	fields := []string{"name", "item.item_name"}
	// "id" is the complusory field.
	// "item_id" is required if "item" would be queried
	relationFields := qf.ApplyFields(fields, []string{"id"}, map[string][]string{ "item": {"item_id"} })
	// Create another query factory for relation
	qf := gormquery.QueryFactory{ Query: models.DbServer.Model(&itemModel{}) }
	itemRelationFields := qf.ApplyFields(relationFields["items"], []string{"item_id"}, map[string][]string{})
*/
func (qf *QueryFactory) ApplyFields(fields []string, compulsoryFields []string, relations map[string][]string) (relationFieldMap map[string][]string) {
	if len(fields) == 0 {
		return
	}
	relationFieldMap = map[string][]string{}
	rootFields := []string{}
	for _, field := range fields {
		isRelation := false
		var relation string
		var dependencies []string
		for relation, dependencies = range relations {
			// exact relation: select relation with all fields
			if field == relation {
				isRelation = true
				relationFieldMap[relation] = []string{}
				break
			}
			// check relational sub-field
			prefix := fmt.Sprintf("%s.", relation)
			if strings.HasPrefix(field, prefix) {
				// add relation
				isRelation = true
				subField := strings.Replace(field, prefix, "", 1)
				if relationFieldMap[relation] == nil {
					relationFieldMap[relation] = []string{subField}
				} else {
					relationFieldMap[relation] = append(relationFieldMap[relation], subField)
				}
				break
			}
		}
		if isRelation {
			rootFields = append(rootFields, dependencies...)
			continue
		}
		// filter out other relational field
		if strings.Contains(field, ".") {
			continue
		}
		// apply field
		rootFields = append(rootFields, field)
	}
	rootFields = append(rootFields, compulsoryFields...)
	qf.Query = qf.Query.Select(rootFields)
	return
}

/*
Apply filter query

# Input

fields: field array

queryObject: primitive or primitive array

# Output

query factory itself for fluent interface pattern

# Example

	// Create query factory
	qf := gormquery.QueryFactory{ Query: models.DbServer.Model(&model{}) }
	// query with simple primitive (e.g. WHERE fieldA = 'stringValue')
	qf.ApplyQuery("fieldA", "stringValue")
	// query with array (e.g. WHERE fieldA in (values...))
	qf.ApplyQuery("fieldB", []string{"stringValue1", "stringValue2"})
*/
func (qf *QueryFactory) ApplyQuery(field string, queryObject any) *QueryFactory {
	if queryObject == nil || field == "" {
		return qf
	}
	switch queryObject := queryObject.(type) {
	case string:
		qf.applyQueryPrimitive(field, queryObject)
	case bool:
		qf.applyQueryPrimitive(field, queryObject)
	case float64:
		qf.applyQueryPrimitive(field, queryObject)
	case int:
		qf.applyQueryPrimitive(field, queryObject)
	// included in
	case []any:
		qf.applyQueryIncludedIn(field, queryObject)
	// TODO: operator
	default:
		fmt.Printf("unknown qf object type of %s", queryObject)
	}
	return qf
}

func (qf *QueryFactory) applyQueryPrimitive(field string, queryValue any) *QueryFactory {
	qf.Query = qf.Query.Where(field, queryValue)
	return qf
}

func (qf *QueryFactory) applyQueryIncludedIn(field string, values []any) *QueryFactory {
	queryTemplate := fmt.Sprintf("%s in ?", field)
	qf.Query = qf.Query.Where(queryTemplate, values)
	return qf
}

// Deprecated: ExtractSimpleJsonQuery is deprecated.
func ExtractSimpleJsonQuery(keyValuePairs []string) (queryMap map[string]string) {
	queryMap = make(map[string]string)
	for _, pair := range keyValuePairs {
		if strings.Contains(pair, ":") {
			sliceStr := strings.Split(pair, ":")
			if len(sliceStr) < 2 {
				continue
			}
			queryMap[sliceStr[0]] = sliceStr[1]
		}
	}
	return
}

// Deprecated: ApplySimpleJsonQuery is deprecated.
/*
Simple format for JSON type field qf

Example:

	Input
		field: "tags"
		keyValuePairs: ["app:APP1", "platform:web"]
	Result Query
		qf.Where("tags ->> ? = ?", "app", "APP1").Where("tags ->> ? = ?", "platform", "web")
*/
func (qf *QueryFactory) ApplySimpleJsonQuery(field string, keyValuePairs []string) *QueryFactory {
	queryMap := ExtractSimpleJsonQuery(keyValuePairs)
	queryTemplate := fmt.Sprintf("%s ->> ? = ?", field)
	for key, value := range queryMap {
		// TODO: apply "included in" for array value
		qf.Query = qf.Query.Where(queryTemplate, key, value)
	}
	return qf
}

func (qf *QueryFactory) ApplySort(field string, sortDef string) *QueryFactory {
	switch sortDef {
	case "DESC":
		qf.Query = qf.Query.Order(fmt.Sprintf("%s DESC", field))
	case "ASC":
		qf.Query = qf.Query.Order(fmt.Sprintf("%s ASC", field))
	default:
		log.Printf("[gormquery] wrong sort def. key: %s, value: %s \n", field, sortDef)
	}
	return qf
}

func ConvertJsonValue(record map[string]any, path string) {
	tarMap := skmap.Map(record)
	v, err := tarMap.Get(path)
	if err != nil {
		return
	}
	bytes, ok := v.(string)
	if !ok {
		return
	}
	tarMap.Set(path, json.RawMessage(bytes))
}

func ConvertJsonValueInRecords(records []map[string]any, paths []string) {
	for _, record := range records {
		for _, path := range paths {
			ConvertJsonValue(record, path)
		}
	}
}

func MapRecordsToStringArray(records []map[string]any) (results []string) {
	// return lo.Map(records, func(record map[string]any, _ int) string {
	// 	bytes, _ := json.Marshal(record)
	// 	return string(bytes)
	// })
	// implement it without lo
	for _, record := range records {
		bytes, _ := json.Marshal(record)
		results = append(results, string(bytes))
	}
	return
}

func ExtractJsonArray(record map[string]any, path string, key string) (success bool) {
	success = false
	if key == "" {
		key = path
	}
	tarMap := skmap.Map(record)
	jsonb, err := tarMap.Get(path)
	if err != nil {
		return
	}
	var bytes []byte
	switch val := jsonb.(type) {
	case []byte:
		bytes = val
	case json.RawMessage:
		bytes = []byte(val)
	case string:
		bytes = []byte(val)
	default:
		return
	}
	obj := skmap.Map{}
	err = json.Unmarshal([]byte(bytes), &obj)
	if err != nil {
		return
	}
	arr := obj.GetArrayDefault(key, nil)
	tarMap.Set(path, arr)
	success = len(arr) > 0
	return
}

func ExtractJsonArrayInRecords(records []map[string]any, path string, key string) (success bool) {
	success = false
	for _, record := range records {
		success = ExtractJsonArray(record, path, key) || success
	}
	return
}
