package utils

import (
	"fmt"
	"strconv"
	"testing"
)

func TestGetShard(t *testing.T) {
	if getShard(3972368057761799, 64, AlgNone) != 7 {
		t.Error("TestGetShard Error")
	}
	if getShard(3972368057761799, 64, AlgCrc32) != 10 {
		t.Error("TestGetShard Error")
	}
	if getShard(3972368057761799, 64, -1) != 3972368057761799 {
		t.Error("TestGetShard Error")
	}

	if getShard(3972368057761799, 100, AlgNone) != 99 {
		t.Error("TestGetShard Error")
	}
	if getShard(3972368057761799, 100, AlgCrc32) != 82 {
		t.Error("TestGetShard Error")
	}
	if getShard(3972368057761799, 100, -1) != 3972368057761799 {
		t.Error("TestGetShard Error")
	}
}
func TestGetShardDb(t *testing.T) {
	if GetShardDb(3972368057761799, 64, AlgCrc32) != getShard(3972368057761799, 64, AlgCrc32) {
		t.Error("TestGetShard Error")
	}

	if GetShardDb(3972368057761799, 100, AlgNone) != getShard(3972368057761799, 100, AlgNone) {
		t.Error("TestGetShard Error")
	}
}
func TestGetShardDbTable(t *testing.T) {
	dbNum, tableNum := GetShardDbTable(3972368057761799, 64, AlgCrc32, 8)
	fmt.Printf("TestGetShardDbTable dbTable: %d,%d\n", dbNum, tableNum)
	dbNum, tableNum = GetShardDbTable(3972368057761799, 100, AlgNone, 10)
	fmt.Printf("TestGetShardDbTable dbTable: %d,%d\n", dbNum, tableNum)
}

func TestGetDbTableName(t *testing.T) {
	if GetDbTableName("testdb", 0, "testtable", 0) != "testdb_0.testtable_0" {
		t.Error("TestGetDbTableName Error")
	}
}
func TestGetDbName(t *testing.T) {
	if GetDbName("testdb", 0) != "testdb_0" {
		t.Error("TestGetDbName Error")
	}
}
func TestGetTableName(t *testing.T) {
	if GetTableName("testtable", 0) != "testtable_0" {
		t.Error("TestGetTableName Error")
	}
}

func TestBuildSql(t *testing.T) {
	if BuildSql("select * from $dbTable$ where id=?", "testdb_0.testtable_0") != "select * from testdb_0.testtable_0 where id=?" {
		t.Error("TestBuildSql Error")
	}
}
func TestBuildSqlDb(t *testing.T) {
	shardNum := GetShardDb(3972368057761799, 64, AlgCrc32)
	if BuildSql("select * from $dbTable$ where id=?", "testdb_"+strconv.Itoa(shardNum)+".testtable") != "select * from testdb_10.testtable where id=?" {
		t.Error("TestBuildSqlDb Error")
	}
}
func TestBuildSqlDbTable(t *testing.T) {
	dbNum, tableNum := GetShardDbTable(3972368057761799, 64, AlgCrc32, 8)
	if BuildSql("select * from $dbTable$ where id=?", GetDbTableName("testdb", dbNum, "testtable", tableNum)) != "select * from testdb_1.testtable_2 where id=?" {
		t.Error("TestBuildSqlDbTable Error")
	}
	if BuildSql("select * from $dbTable$ where id=?", GetTableName("testtable", tableNum)) != "select * from testtable_2 where id=?" {
		t.Error("TestBuildSqlDbTable Error")
	}
}
func TestGetShardSql(t *testing.T) {
	if GetShardSql(3972368057761799, 64, AlgCrc32, 8, "select * from $dbTable$ where id=?", "testdb", "testtable") != "select * from testdb_1.testtable_2 where id=?" {
		t.Error("TestGetShardSql Error")
	}
	if GetShardSql(3972368057761799, 100, AlgNone, 10, "select * from $dbTable$ where id=?", "testdb", "testtable") != "select * from testdb_9.testtable_9 where id=?" {
		t.Error("TestGetShardSql Error")
	}
}
