package xcrypto

import (
	"crypto/cipher"
	"reflect"
	"testing"
)

func TestNewDes(t *testing.T) {
	type args struct {
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "new1",
			args: args{
				key: []byte("hahahehe"),
			},
			wantErr: false,
		},
		{
			name: "new2",
			args: args{
				key: []byte("hahahehexixi"),
			},
			wantErr: true,
		},
		{
			name: "new3",
			args: args{
				key: []byte("haha"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDes(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDesCrypter_Encrypt(t *testing.T) {
	des, _ := NewDes([]byte("hahahehe"))
	type fields struct {
		cb  cipher.Block
		key []byte
		bs  int
	}
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "en1",
			fields: fields(*des),
			args: args{
				src: []byte("12345678"),
			},
			want: []byte{60, 246, 96, 234, 219, 81, 247, 67, 172, 98, 159, 141, 49, 163, 40, 242},
		},
		{
			name:   "en2",
			fields: fields(*des),
			args: args{
				src: []byte("12345678910"),
			},
			want: []byte{60, 246, 96, 234, 219, 81, 247, 67, 40, 21, 37, 22, 105, 79, 141, 180},
		},
		{
			name:   "en3",
			fields: fields(*des),
			args: args{
				src: []byte(""),
			},
			want: []byte{56, 18, 196, 229, 208, 235, 171, 157},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &DesCrypter{
				cb:  tt.fields.cb,
				bs:  tt.fields.bs,
				key: tt.fields.key,
			}
			got, err := this.Encrypt(tt.args.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DesCrypter.Encrypt() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DesCrypter.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDesCrypter_Decrypt(t *testing.T) {
	des, _ := NewDes([]byte("hahahehe"))
	type fields struct {
		cb  cipher.Block
		key []byte
		bs  int
	}
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "de1",
			fields: fields(*des),
			args: args{
				src: []byte{60, 246, 96, 234, 219, 81, 247, 67, 172, 98, 159, 141, 49, 163, 40, 242},
			},
			want: []byte("12345678"),
		},
		{
			name:   "de2",
			fields: fields(*des),
			args: args{
				src: []byte{60, 246, 96, 234, 219, 81, 247, 67, 40, 21, 37, 22, 105, 79, 141, 180},
			},
			want: []byte("12345678910"),
		},
		{
			name:   "de3",
			fields: fields(*des),
			args: args{
				src: []byte{56, 18, 196, 229, 208, 235, 171, 157},
			},
			want: []byte(""),
		},
		{
			name:   "de4",
			fields: fields(*des),
			args: args{
				src: []byte{},
			},
			want: []byte(""),
		},
		{
			name:   "de5",
			fields: fields(*des),
			args: args{
				src: []byte{56, 18, 196, 229, 208, 235, 171},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "de6",
			fields: fields(*des),
			args: args{
				src: []byte{56, 18, 196, 229, 208, 235, 171, 158},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &DesCrypter{
				bs:  tt.fields.bs,
				cb:  tt.fields.cb,
				key: tt.fields.key,
			}
			got, err := this.Decrypt(tt.args.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DesCrypter.Decrypt() = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("DesCrypter.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
