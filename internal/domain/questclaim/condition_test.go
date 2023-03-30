package questclaim

import (
	"reflect"
	"testing"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/testutil"
)

func Test_newQuestCondition(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)
	questRepo := repository.NewQuestRepository(db)

	type args struct {
		condition entity.Condition
	}

	tests := []struct {
		name    string
		args    args
		want    *questCondition
		wantErr bool
	}{
		{
			name: "happy case with is completed",
			args: args{
				condition: entity.Condition{
					Op:    string(isCompleted),
					Value: testutil.Quest1.ID,
				}},
			want: &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               isCompleted,
				questID:          testutil.Quest1.ID,
			},
			wantErr: false,
		},
		{
			name: "happy case with is not completed",
			args: args{
				condition: entity.Condition{
					Op:    string(isNotCompleted),
					Value: testutil.Quest2.ID,
				}},
			want: &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               isNotCompleted,
				questID:          testutil.Quest2.ID,
			},
			wantErr: false,
		},
		{
			name: "invalid op",
			args: args{
				condition: entity.Condition{
					Op:    "invalid op",
					Value: testutil.Quest2.ID,
				}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid quest id",
			args: args{
				condition: entity.Condition{
					Op:    string(isCompleted),
					Value: "invalid quest id",
				}},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newQuestCondition(
				testutil.NewMockContext(),
				tt.args.condition,
				claimedQuestRepo,
				questRepo,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("newQuestCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newQuestCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_questCondition_Check(t *testing.T) {
	db := testutil.CreateFixtureDb()
	claimedQuestRepo := repository.NewClaimedQuestRepository(db)

	type fields struct {
		op      questConditionOpType
		questID string
	}

	type args struct {
		ctx router.Context
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "happy case with is completed",
			fields: fields{
				op:      isCompleted,
				questID: testutil.Quest1.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.ClaimedQuest1.UserID),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "happy case with is not completed",
			fields: fields{
				op:      isCompleted,
				questID: testutil.Quest2.ID,
			},
			args: args{
				ctx: testutil.NewMockContextWithUserID(testutil.ClaimedQuest1.UserID),
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &questCondition{
				claimedQuestRepo: claimedQuestRepo,
				op:               tt.fields.op,
				questID:          tt.fields.questID,
			}

			got, err := c.Check(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("questCondition.Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("questCondition.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newDateCondition(t *testing.T) {
	type args struct {
		condition entity.Condition
	}
	tests := []struct {
		name    string
		args    args
		want    *dateCondition
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				condition: entity.Condition{
					Op:    "before",
					Value: "Mar 29 2023",
				},
			},
			want: &dateCondition{
				op:   dateBefore,
				date: time.Date(2023, time.March, 29, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "invalid op",
			args: args{
				condition: entity.Condition{
					Op:    "invalid op",
					Value: "Mar 29 2023",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid date",
			args: args{
				condition: entity.Condition{
					Op:    "after",
					Value: "29 Mar 2023",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDateCondition(testutil.NewMockContext(), tt.args.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dateCondition_Check(t *testing.T) {
	type fields struct {
		op   dateConditionOpType
		date time.Time
	}

	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		{
			name: "happy case with date before",
			fields: fields{
				op:   dateBefore,
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "happy case with date after",
			fields: fields{
				op:   dateAfter,
				date: time.Now().AddDate(0, 0, -1),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failed with date before",
			fields: fields{
				op:   dateBefore,
				date: time.Now().AddDate(0, 0, -1),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "failed with date after",
			fields: fields{
				op:   dateAfter,
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "invalid op",
			fields: fields{
				op:   "invalid op",
				date: time.Now().AddDate(0, 0, 1),
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &dateCondition{
				op:   tt.fields.op,
				date: tt.fields.date,
			}

			got, err := c.Check(testutil.NewMockContext())
			if (err != nil) != tt.wantErr {
				t.Errorf("dateCondition.Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("dateCondition.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
