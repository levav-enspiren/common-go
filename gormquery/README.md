# gORM Query Utilities

## gormquery

A query factory to construct gORM query

Example
``` go
	// Create query factory
	qf := gormquery.QueryFactory{ Query: models.DbServer.WithContext(ctx).Model(&model{}) }
	// To select normal field "name" and relational field "item",
	// with item field selection "item_name"
	fields := []string{"name", "item.item_name"}
	// "id" is the complusory field.
	// "item_id" is required if "item" would be queried
	relationFields := qf.ApplyFields(fields, []string{"id"}, map[string][]string{ "item": {"item_id"} })
	// Create another query factory for relation
	qf := gormquery.QueryFactory{ Query: models.DbServer.Model(&itemModel{}) }
	itemRelationFields := qf.ApplyFields(relationFields["items"], []string{"item_id"}, map[string][]string{})
  // apply "where" query
  qf.ApplyQuery("fieldName0", primitiveValue).                    // simple equal statement
    ApplyQuery("fieldName1", []string{["value0", "value1"]}).     // included in statement
    ApplyQuery("fieldName2", skmap.Map{ "$gt": 123 })             // TODO: complex query
  // get the result query
  query := qf.Query
  // perform gORM operation
  err := query.Find(&results).Error
```

## grpc-service

A gRPC service server to be registerred to gRPC server

### Setup
``` go
  // setup after rpcServer is ready
	queryService.RegisterQueryServiceServer(rpcServer, &gormquery.QueryServiceServer{
		ModelClasses: map[string]gormquery.ModelClass{
			"item": {
				Model:                 postgres.Item{},
				CanGet:                true,
				WhitelistedFields: skmap.Map{
					"item_id": true,
					"tags": skmap.Map{
						"isJsonField": true,
					},
					"context_id":      true,
					"context_variant": true,
					"obsoleted":       true,
				},
				QueryComplusoryFields: []string{"id"},
			},
		},
		DefaultModelClass: "item",
		DefaultDb:         postgres.DbServer,
	})
```

More about gormquery.ModelClass:
- Db: the database of the model. DefaultDb would be used if it is not provided
- CanGet, CanUpdate, CanCreate, CanDelete: flag to control accessability of API

### gRPC API
- rpc Get(OptionRequest) returns (QueryResponse){};
- rpc Create(OptionRequest) returns (CreateResponse){};
- rpc Update(OptionRequest) returns (Empty){};
- rpc Delete(OptionRequest) returns (Empty){};

#### OptionRequest
- bytes options: to be unmarshaled as string-key-map
  - []string fields: selecting field names
  - integer limit: page size
  - ineger page: page index
  - string keyword: TODO: for searching
  - map sorter: TODO: for sorting
  - map filter: to construct "where" query

## gprc-client-model

A gRPC service client, and model to access the API

### Setup
``` go
// as a model root variable
var QueryServiceModel gormquery.QueryServiceModel

// setup after grpc connection is created
	queryServiceClient = queryService.NewQueryServiceClient(grpcConn)
	QueryServiceModel = gormquery.QueryServiceModel{
		GrpcClient:     queryServiceClient,
		RequestTimeout: time.Duration(config.GrpcConfig.CallTimeout),
	}
```

### Usage
``` go
  // extract options from router context
	options := util.ExtractQueryOption(ctx)
	// skip controller for typical CURD
	results, totalCount, err := auditTrailModel.QueryServiceModel.Get(rCtx, options)

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
```
