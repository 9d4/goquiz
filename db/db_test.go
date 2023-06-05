package db

import (
	"testing"
)

func Test_connect(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "connect db goquiz.test.sqlite3",
			args:    args{name: "goquiz.test.sqlite3"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := connect(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
