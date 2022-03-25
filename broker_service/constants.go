package main

const (
	FAMILY_NAME string = "broker"
	FAMILY_VERSION string = "1.0"
	DISTRIBUTION_NAME string = "sawtooth-broker"
	DEFAULT_URL string = "http://127.0.0.1:8008"

	BATCH_SUBMIT_API string = "batches"
	BATCH_STATUS_API string = "batch_statuses"
	STATE_API string = "state"

	CONTENT_TYPE_OCTET_STREAM string = "application/octet-stream"

	SAWTOOTH_URL = "http://127.0.0.1:8008"
	//PREFIX = "19d832"
	DATA_NAMESPACE = "19d832"
	META_NAMESPACE = "19d732"
	FAMILY_NAMESPACE_ADDRESS_LENGTH uint = 6
	FAMILY_VERB_ADDRESS_LENGTH uint = 62
)
