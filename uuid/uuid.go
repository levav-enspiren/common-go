package uuid

import (
	guuid "github.com/google/uuid"
)

func NewV4Uuid() string {
	u4 := guuid.New()
	return u4.String()
}

func ParseUuidTimestamp(targetUuid string) (int64, error) {
	parseUuid, err := guuid.Parse(targetUuid)
	if err != nil {
		return 0, err
	}
	sec, _ := parseUuid.Time().UnixTime()
	return sec, nil
}

func NewV1Uuid() (string, error) {
	newUuid, err := guuid.NewUUID()
	return newUuid.String(), err
}
