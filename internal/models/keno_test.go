package models

import (
	"testing"

	"github.com/lib/pq"
)

func TestKeno_BeforeSave(t *testing.T) {
	tests := []struct {
		name    string
		ketQua  pq.Int64Array
		wantErr bool
	}{
		{
			name:    "Valid KetQua with 5 elements",
			ketQua:  pq.Int64Array{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "Invalid KetQua with 4 elements",
			ketQua:  pq.Int64Array{1, 2, 3, 4},
			wantErr: true,
		},
		{
			name:    "Invalid KetQua with 6 elements",
			ketQua:  pq.Int64Array{1, 2, 3, 4, 5, 6},
			wantErr: true,
		},
		{
			name:    "Empty KetQua",
			ketQua:  pq.Int64Array{},
			wantErr: true,
		},
		{
			name:    "Nil KetQua",
			ketQua:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Keno{
				KetQua: tt.ketQua,
			}
			err := k.BeforeSave(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeforeSave() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
