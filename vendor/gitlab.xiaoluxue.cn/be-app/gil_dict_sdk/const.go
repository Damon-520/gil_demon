package gil_dict_sdk

const (
	HTTP_GET_META_DICTS_PATH   = "/api/v1/internal/dicts"
	HTTP_GET_ALL_REGIONS_PATH  = "/api/v1/internal/listAllProvinces"
	HTTP_GET_ALL_PROVINCE_PATH = "/api/v1/internal/listAllTopProvinces"

	CACHE_KEY_GET_ALL_REGION        = "admin:dict:region:all"
	CACHE_KEY_GET_ALL_PROVINCE      = "admin:dict:region:all:province"
	CACHE_KEY_GET_ALL_META_DICT     = "admin:dict:meta:sys:all"
	CACHE_KEY_GET_META_DICT_BY_TYPE = "admin:dict:meta:%s:%s"
)
