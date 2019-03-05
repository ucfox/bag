package redis

import "testing"
import "math"

func Test_indexInt32_Incr(t *testing.T) {
	incr2 := int32(math.MaxInt32)
	tests := []struct {
		name string
		this *indexInt32
		want int
	}{
		{
			name: "incr1",
			this: new(indexInt32),
			want: 1,
		},
		{
			name: "incr2",
			this: (*indexInt32)(&incr2),
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Incr(); got != tt.want {
				t.Errorf("indexInt32.Incr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_indexInt32_IncrAndMod(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		this *indexInt32
		args args
		want int
	}{
		{
			name: "incrAndMod1",
			this: new(indexInt32),
			args: args{
				n: 10,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.IncrAndMod(tt.args.n); got != tt.want {
				t.Errorf("indexInt32.IncrAndMod() = %v, want %v", got, tt.want)
			}
		})
	}
}
