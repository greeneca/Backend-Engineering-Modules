package data

import (
	"testing"
	mock_stores "wiki_updates/data/stores/mock"
	"wiki_updates/models"
	"wiki_updates/test_utils"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Cassandra_createTables(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name string
		expectations func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface)
		panics bool
	}{
		{
			name: "Test createTables executes queries",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(gomock.Any()).Return(q).Times(4)
				q.EXPECT().Exec().Return(nil).Times(4)
			},
			panics: false,
		},
		{
			name: "Test createTables panics on query error",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(gomock.Any()).Return(q).Times(1)
				q.EXPECT().Exec().Return(assert.AnError).Times(1)
			},
			panics: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := mock_stores.NewMockSessionInterface(ctrl)
			mock_query := mock_stores.NewMockQueryInterface(ctrl)
			tt.expectations(mockSession, mock_query)
			if tt.panics {
				assert.Panics(t, func() {createTables(mockSession)})
				return
			} else {
				createTables(mockSession)
			}
		})
	}
}

func Test_Cassandra_SaveUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name string
		update models.Update
		expectations func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface)
		errorExpected bool
	}{
		{
			name: "Test SaveUpdate executes insert query",
			update: models.Update{
				Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
				User: "BotUser",
				Bot:  true,
			},
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO wiki_users .*`), "BotUser", true).Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT id FROM wiki_users .*`), "BotUser", true).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO uris .*`), "https://en.wikipedia.org/wiki/Special:Diff/1234567890").Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT id FROM uris .*`), gomock.Any()).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO updates .*`), gomock.Any(), gomock.Any()).Return(q)
				q.EXPECT().Exec().Return(nil)
			},
			errorExpected: false,
		},
		{
			name: "Test SaveUpdate returns error on insert wiki_user failure",
			update: models.Update{
				Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
				User: "BotUser",
				Bot:  true,
			},
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO wiki_users .*`), "BotUser", true).Return(q)
				q.EXPECT().Exec().Return(assert.AnError)
			},
			errorExpected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := mock_stores.NewMockSessionInterface(ctrl)
			mock_query := mock_stores.NewMockQueryInterface(ctrl)
			tt.expectations(mockSession, mock_query)
			db := &Cassandra{session: mockSession}
			err := db.SaveUpdate(tt.update)
			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Cassandra_GetStatistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name string
		expectations func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface)
		expected models.Statistics
	}{
		{
			name: "Test GetStatistics returns correct statistics",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM updates`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
					*arg = 100
					return nil
				})
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM uris`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
					*arg = 50
					return nil
				})
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = true`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
					*arg = 30
					return nil
				})
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = false`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
					*arg = 70
					return nil
				})
			},
			expected: models.Statistics{
				Messages: 100,
				Urls:     50,
				Bots:     30,
				NonBots:  70,
			},
		},
		{
			name: "Test GetStatistics handles query errors gracefully",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM updates`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM uris`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = true`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = false`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
			},
			expected: models.Statistics{
				Messages: 0,
				Urls:     0,
				Bots:     0,
				NonBots:  0,
			},
		},
		{
			name: "Test GetStatistics does not return early on errors",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM updates`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM uris`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = true`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT COUNT\(\*\) FROM wiki_users WHERE bot = false`)).Return(q)
				q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
					*arg = 70
					return nil
				})
			},
			expected: models.Statistics{
				Messages: 0,
				Urls:     0,
				Bots:     0,
				NonBots:  70,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := mock_stores.NewMockSessionInterface(ctrl)
			mock_query := mock_stores.NewMockQueryInterface(ctrl)
			tt.expectations(mockSession, mock_query)
			db := &Cassandra{
				session: mockSession,
				stats:   models.Statistics{},
			}
			stats, err := db.GetStatistics()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, *stats)
		})
	}
}

func Test_Cassandra_GetUserByEmail(t *testing.T) {
	email := "test@user.com"
	ctrl := gomock.NewController(t)
	session := mock_stores.NewMockSessionInterface(ctrl)
	query := mock_stores.NewMockQueryInterface(ctrl)
	db := &Cassandra{session: session}
	session.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT email, password_hash FROM users .*`), email).Return(query).Times(1)
	query.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...*string) error {
		*dest[0] = email
		*dest[1] = "hashed_password"
		return nil
	}).Times(1)
	user,  err := db.GetUserByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
}

func Test_Cassandra_SaveUser(t *testing.T) {
	test_user := &models.User{
		Email: "test@user.com",
		PasswordHash: "hashed_password",
	}
	ctrl := gomock.NewController(t)
	session := mock_stores.NewMockSessionInterface(ctrl)
	query := mock_stores.NewMockQueryInterface(ctrl)
	db := &Cassandra{session: session}
	session.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT email, password_hash FROM users .*`), test_user.Email) .Return(query).Times(1)
	query.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(assert.AnError).Times(1)
	session.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO users .*`), gomock.Any(), gomock.Any()).Return(query).Times(1)
	query.EXPECT().Exec().Return(nil).Times(1)
	err := db.SaveUser(test_user)
	assert.NoError(t, err)
}

