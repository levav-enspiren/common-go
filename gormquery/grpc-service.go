package gormquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"gitea.greatics.net/common-go/gormquery/helper"
	"gitea.greatics.net/common-go/gormquery/queryService"
	"gitea.greatics.net/common-go/skmap"
	"gorm.io/gorm"
)

type ModelClass struct {
	Model any
	Db    *gorm.DB
	// Get Config
	CanGet                bool
	WhitelistedFields     skmap.Map
	QueryComplusoryFields []string
	// TODO:
	// queryRelations        []ModelRelation
	// Create Config
	CanCreate bool
	// Update Config
	CanUpdate bool
	// Delete Config
	CanDelete bool
}

func (mc *ModelClass) CreateModelRef() any {
	return reflect.Zero(reflect.TypeOf(mc.Model)).Interface()
}

func (mc *ModelClass) CreateModelArrayRef() any {
	return reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(mc.Model)), 0, 0).Interface()
}

type QueryServiceServer struct {
	queryService.QueryServiceServer

	ModelClasses      map[string]ModelClass
	DefaultModelClass string
	DefaultDb         *gorm.DB
}

func (q *QueryServiceServer) parseOptionRequest(request *queryService.OptionRequest) (options skmap.Map, modelClass ModelClass, db *gorm.DB, err error) {
	err = json.Unmarshal(request.Options, &options)
	if err != nil {
		return
	}
	//
	modelClassName := options.GetStringDefault("modelClass", q.DefaultModelClass)
	modelClass, ok := q.ModelClasses[modelClassName]
	if !ok {
		err = errors.New("invalid model class")
		return
	}
	//
	db = modelClass.Db
	if db == nil {
		db = q.DefaultDb
	}
	if db == nil {
		err = errors.New("missing db")
		return
	}
	return
}

func applyFilter(qf QueryFactory, filter skmap.Map, whitelistedFields skmap.Map) (hasFilter bool) {
	hasFilter = false
	if filter == nil {
		return
	}
	for field, _ := range whitelistedFields {
		isJsonField := whitelistedFields.GetBoolDefault(fmt.Sprintf("%s.isJsonField", field), false)
		if isJsonField {
			filterValue := filter.GetStringArraySafe(field)
			if len(filterValue) == 0 {
				continue
			}
			hasFilter = true
			qf.ApplySimpleJsonQuery(field, filterValue)
		} else {
			filterValue := filter.GetDefault(field, nil)
			if filterValue == nil {
				continue
			}
			hasFilter = true
			qf.ApplyQuery(field, filterValue)
		}
	}
	return
}

func applySortDefs(qf QueryFactory, sortDefs skmap.Map) {
	for sortField, _ := range sortDefs {
		sortDef := sortDefs.GetStringDefault(sortField, "ASC")
		qf.ApplySort(sortField, sortDef)
	}
}

/*
Perform "Read" operation

# Input

options: JSON string
  - fields []string : selected fields
  - page int : page number
  - limit int : page size
  - keyword string : keyword to search (not implemented)
  - sorter map : sorter description
  - filter map : filter query (ref: gormquery.applyFilter)

# Example

	// example of option extractor for GET call
	func ExtractQueryOption(ctx *gin.Context) (options skmap.Map) {
		fieldStr := ctx.Query("field")
		var fields []string
		if fieldStr != "" {
			fields = strings.Split(fieldStr, ",")
		}
		page, _ := strconv.Atoi(ctx.DefaultQuery("page", "0"))
		limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "0"))
		keyword := ctx.DefaultQuery("keyword", "")
		var sorter skmap.Map
		sorterBytes := ctx.DefaultQuery("sorter", "{}")
		json.Unmarshal([]byte(sorterBytes), &sorter)
		var filter skmap.Map
		filterBytes := ctx.DefaultQuery("filter", "{}")
		json.Unmarshal([]byte(filterBytes), &filter)
		options = skmap.Map{
			"fields":  fields,
			"page":    page,
			"limit":   limit,
			"keyword": keyword,
			"sorter":  sorter,
			"filter":  filter,
		}
		return
	}
*/
func (q *QueryServiceServer) Get(ctx context.Context, request *queryService.OptionRequest) (response *queryService.QueryResponse, err error) {
	options, modelClass, db, err := q.parseOptionRequest(request)
	if err != nil {
		return
	}
	if !modelClass.CanGet {
		err = errors.New("permission denied")
		return
	}
	// transform params
	fields := options.GetStringArraySafe("fields")
	limit := options.GetIntDefault("limit", 0)
	if limit < 0 {
		limit = 0
	}
	page := options.GetIntDefault("page", 0)
	if page < 0 {
		page = 0
	}
	filter := options.GetMapDefault("filter", skmap.Map{})
	sortDefs := options.GetMapDefault("sort", skmap.Map{})
	// construct query
	qf := QueryFactory{Query: db.WithContext(ctx).Model(modelClass.CreateModelRef())}
	// apply fields
	qf.ApplyFields(fields, modelClass.QueryComplusoryFields, map[string][]string{})
	applyFilter(qf, filter, modelClass.WhitelistedFields)
	applySortDefs(qf, sortDefs)
	// get query
	query := qf.Query
	// count
	var totalCountInt int64
	err = query.Count(&totalCountInt).Error
	if err != nil {
		return
	}
	totalCount := uint64(totalCountInt)
	// pagination
	if limit != 0 {
		query = query.Limit(limit).Offset(limit * page)
	}
	// var results []skmap.Hash
	results := modelClass.CreateModelArrayRef()
	err = query.Find(&results).Error
	if err != nil {
		return
	}
	resultsBytes, err := json.Marshal(results)
	if err != nil {
		return
	}
	// return
	response = &queryService.QueryResponse{
		TotalCount: totalCount,
		Results:    resultsBytes,
	}
	return
}

/*
Perform "Create" operation

# Input

options: JSON string
  - data map : record data
*/
func (q *QueryServiceServer) Create(ctx context.Context, request *queryService.OptionRequest) (response *queryService.CreateResponse, err error) {
	options, modelClass, db, err := q.parseOptionRequest(request)
	if err != nil {
		return
	}
	if !modelClass.CanCreate {
		err = errors.New("permission denied")
		return
	}
	data := options.GetMapDefault("data", nil)
	if data == nil {
		err = errors.New("missing data")
		return
	}
	dataHash := helper.CastDataMap(data)
	err = db.WithContext(ctx).Model(modelClass.CreateModelRef()).Create(&dataHash).Error
	if err != nil {
		return
	}
	dataBytes, err := json.Marshal(dataHash)
	response = &queryService.CreateResponse{
		Result: dataBytes,
	}
	return
}

/*
Perform "Update" operation

# Input

options: JSON string
  - data map : record data
  - filter map : filter query (ref: gormquery.applyFilter)
*/
func (q *QueryServiceServer) Update(ctx context.Context, request *queryService.OptionRequest) (response *queryService.Empty, err error) {
	options, modelClass, db, err := q.parseOptionRequest(request)
	if err != nil {
		return
	}
	if !modelClass.CanUpdate {
		err = errors.New("permission denied")
		return
	}
	filter := options.GetMapDefault("filter", nil)
	if filter == nil {
		err = errors.New("missing filter")
		return
	}
	data := options.GetMapDefault("data", nil)
	if data == nil {
		err = errors.New("missing data")
		return
	}
	dataHash := helper.CastDataMap(data)
	// construct query
	qf := QueryFactory{Query: db.WithContext(ctx).Model(modelClass.CreateModelRef())}
	// apply filter
	hasFilter := applyFilter(qf, filter, modelClass.WhitelistedFields)
	if !hasFilter {
		err = errors.New("missing filter")
		return
	}
	// get query
	query := qf.Query
	// update
	err = query.Updates(dataHash).Error
	response = &queryService.Empty{}
	return
}

/*
Perform "Delete" operation

# Input

options: JSON string
  - filter map : filter query (ref: gormquery.applyFilter)
*/
func (q *QueryServiceServer) Delete(ctx context.Context, request *queryService.OptionRequest) (response *queryService.Empty, err error) {
	options, modelClass, db, err := q.parseOptionRequest(request)
	if err != nil {
		return
	}
	if !modelClass.CanDelete {
		err = errors.New("permission denied")
		return
	}
	filter := options.GetMapDefault("filter", nil)
	if filter == nil {
		err = errors.New("missing filter")
		return
	}
	// construct query
	qf := QueryFactory{Query: db.WithContext(ctx)}
	// apply filter
	hasFilter := applyFilter(qf, filter, modelClass.WhitelistedFields)
	if !hasFilter {
		err = errors.New("missing filter")
		return
	}
	//
	query := qf.Query
	//
	err = query.Delete(modelClass.CreateModelRef()).Error
	response = &queryService.Empty{}
	return
}
