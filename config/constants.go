package config

import "github.com/golang/protobuf/jsonpb"

const (
	DateTimeFormat              = "2006-01-02 15:04:05"
	DateFormat                  = "2006-01-02"
	DateTimeFormatWithoutSpaces = "20060102150405"
	ConsumerGroupID             = "invan_catalog_service"

	// product types
	SimpleProductTypeID  = "8b0bf29c-58e8-4310-8bb1-a1b9771f9c47"
	ServiceProductTypeID = "2b98f424-91c9-46cc-abd7-c888208807da"
	SetProductTypeID     = "a19a514e-41c9-4666-a01a-e3f9c0255609"

	// custom field type
	BooleanCFTypeID = "8b0bf29c-58e8-4310-8bb1-a1b9771f9c47"
	StringCFTypeID  = "2b98f424-91c9-46cc-abd7-c888208807da"
	NumberCFTypeID  = "a19a514e-41c9-4666-a01a-e3f9c0255609"

	ElasticProductIndex = "products"

	// DebugMode indicates service mode is debug.
	DebugMode = "debug"
	// TestMode indicates service mode is test.
	TestMode = "test"
	// ReleaseMode indicates service mode is release.
	ReleaseMode = "release"

	FileBucketName = "file"
)

var (
	JSONPBMarshaler = jsonpb.Marshaler{EmitDefaults: true, OrigName: true}
)
