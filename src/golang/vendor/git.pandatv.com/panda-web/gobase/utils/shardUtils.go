package utils

import (
	"strconv"
	"strings"
)

const (
	AlgNone  = 0
	AlgCrc32 = iota
)

func getShard(id uint64, count uint64, alg int) int {
	num := id
	if alg == AlgNone {
		num = id % count
	} else if alg == AlgCrc32 {
		str := strconv.FormatUint(id, 10)
		num = uint64(Crc32(str)) % count
	}
	return int(num)
}
func GetShardDbTable(id uint64, count uint64, alg int, tablePerDb int) (int, int) {
	num := getShard(id, count, alg)
	dbNum := num / tablePerDb
	tableNum := num % tablePerDb
	return dbNum, tableNum
}
func GetShardDb(id uint64, count uint64, alg int) int {
	dbNum, _ := GetShardDbTable(id, count, alg, 1)
	return dbNum
}

func GetDbTableName(db string, dbNum int, table string, tableNum int) string {
	return db + "_" + strconv.Itoa(dbNum) + "." + table + "_" + strconv.Itoa(tableNum)
}
func GetDbName(db string, dbNum int) string {
	return db + "_" + strconv.Itoa(dbNum)
}
func GetTableName(table string, tableNum int) string {
	return table + "_" + strconv.Itoa(tableNum)
}

func BuildSql(rawSql string, dbTable string) string {
	result := strings.Replace(rawSql, "$dbTable$", dbTable, -1)
	return result
}

func GetShardSql(id uint64, count uint64, alg int, tablePerDb int, rawSql string, db string, table string) string {
	dbNum, tableNum := GetShardDbTable(id, count, alg, tablePerDb)
	dbTable := db + "_" + strconv.Itoa(dbNum) + "." + table + "_" + strconv.Itoa(tableNum)
	return BuildSql(rawSql, dbTable)
}
