package logkit

import "testing"

var msg = "fsfsfsdhksdjfkjsdkjfsdkjffhsdkjfhsdjkhfkjsdfhjksuhhkhjhjkfsfsdhksdjfkjsdkjfsdkjffhsdkjfhsdjkhfkjsdfhjksuhhkhjhjkfs\n"

func BenchmarkSyslog(b *testing.B) {
	w := newSyslogWriter("syslog_benchmark_test")
	for i := 0; i < b.N; i++ {
		w.write(LevelInfo, msg)
	}
	w.exit()
}

func BenchmarkFilelog(b *testing.B) {
	w := newFileLog("filelog_benchmark_test", "/data/projlogs/")
	for i := 0; i < b.N; i++ {
		w.write(LevelInfo, msg)
	}
	w.exit()
}

func BenchmarkPipeLog(b *testing.B) {
	w := newPipelogWriter("pipelog_benchmark_test", "/data/projlogs/golanglog")
	for i := 0; i < b.N; i++ {
		w.write(LevelInfo, msg)
	}
	w.exit()
}
