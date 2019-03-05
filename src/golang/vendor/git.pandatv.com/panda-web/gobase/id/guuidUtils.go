package id

import ()

const (
	ID_OFFSET               = 1251414000
	SEQ_BIT_LENGTH          = 18
	INSTANCE_SEQ_BIT_LENGTH = 6 + SEQ_BIT_LENGTH
)

func GetTimeFromUuid(id uint64) uint64 {
	time := (id >> INSTANCE_SEQ_BIT_LENGTH) + ID_OFFSET
	return time
}
