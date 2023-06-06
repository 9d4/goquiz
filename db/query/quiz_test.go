package query

import (
	"reflect"
	"testing"
)

func TestNewChoice(t *testing.T) {
	type args struct {
		body       string
		questionID int
		correct    []bool
	}
	tests := []struct {
		name string
		args args
		want *Choice
	}{
		{
			name: "New Choice 0",
			args: args{
				body:       "Choice 0",
				questionID: 1,
				correct:    nil,
			},
			want: &Choice{
				Body:       "Choice 0",
				QuestionID: 1,
				Correct:    false,
			},
		},

		{
			name: "New Choice 1",
			args: args{
				body:       "Choice 1",
				questionID: 1,
				correct:    []bool{true},
			},
			want: &Choice{
				Body:       "Choice 1",
				QuestionID: 1,
				Correct:    true,
			},
		},

		{
			name: "New Choice 2",
			args: args{
				body:       "Choice 2",
				questionID: 1,
				correct:    []bool{false},
			},
			want: &Choice{
				Body:       "Choice 2",
				QuestionID: 1,
				Correct:    false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChoice(tt.args.body, tt.args.questionID, tt.args.correct...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewChoice() = %v, want %v", got, tt.want)
			}
		})
	}
}
