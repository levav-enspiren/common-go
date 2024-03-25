package helper

import "gitea.greatics.net/common-go/skmap"

func CastDataMap(data skmap.Map) skmap.Hash {
	for field := range data {
		if mapData, _ := data[field].(skmap.Hash); mapData != nil {
			data[field] = skmap.Map(mapData)
		}
	}
	return skmap.Hash(data)
}
