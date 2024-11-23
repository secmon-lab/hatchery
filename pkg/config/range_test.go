package config_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/secmon-lab/hatchery/pkg/config"
	"github.com/urfave/cli/v3"
)

func TestRange(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		minutes  []int
		expected int
	}{
		{
			name: "Specify start and end time",
			options: []string{
				"-s", "2024-11-20T00:00:00Z",
				"-e", "2024-11-20T00:05:00Z",
				"-d", "2m",
			},
			minutes:  []int{0, 2, 4},
			expected: 3,
		},
		{
			name:     "no option",
			options:  []string{},
			minutes:  []int{},
			expected: 1,
		},
		{
			name: "Specify start time only",
			options: []string{
				"-s", "2024-11-20T00:00:00Z",
			},
			minutes:  []int{0},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rangeCfg config.Range
			var cnt int

			app := cli.Command{
				Flags: rangeCfg.Flags(),
				Action: func(ctx context.Context, c *cli.Command) error {
					for v := range rangeCfg.Generate {
						cnt++
						if len(tt.minutes) > 0 {
							gt.Equal(t, v.Minute(), tt.minutes[cnt-1])
						}
					}
					return nil
				},
			}

			gt.NoError(t, app.Run(context.Background(), append([]string{"app"}, tt.options...)))
			gt.Equal(t, cnt, tt.expected)
		})
	}
}
