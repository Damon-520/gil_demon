package consts

import "sort"

type Dict map[int]string

func (d Dict) GetValue(key int) string {
	return d[key]
}

type DictKv struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// ToSlice 将字典转换为切片，默认按照id升序排序
func (d Dict) ToSlice() []DictKv {

	result := make([]DictKv, 0, len(d))
	for k, v := range d {
		result = append(result, DictKv{
			Id:   k,
			Name: v,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})

	return result
}

func GetDictByKey(key string) Dict {
	switch key {

	}

	return Dict{} // 默认返回空字典，返回nil可能会导致panic
}
