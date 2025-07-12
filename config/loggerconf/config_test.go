package loggerconf

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		conf    Config
		wantErr bool
	}{
		{
			name: "valid config with file",
			conf: Config{
				Outputs: []string{"file"},
				FileOutput: &FileOutputConfig{
					Path: "/tmp/test.log",
				},
				Backend: "zerolog",
			},
			wantErr: false,
		},
		{
			name: "file output missing path",
			conf: Config{
				Outputs:    []string{"file"},
				FileOutput: &FileOutputConfig{}, // missing Path
				Backend:    "zap",
			},
			wantErr: true,
		},
		{
			name: "invalid backend",
			conf: Config{
				Outputs: []string{"stdout"},
				Backend: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conf.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
