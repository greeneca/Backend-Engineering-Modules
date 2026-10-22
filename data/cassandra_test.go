package data

import (
	"testing"
	mock_stores "wiki_updates/data/stores/mock"
	"wiki_updates/models"
	"wiki_updates/test_utils"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
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
	update := models.Update{
		Uri:  "https://en.wikipedia.org/wiki/Special:Diff/1234567890",
		User: "BotUser",
		Bot:  true,
	}
	tests := []struct {
		name string
		update models.Update
		expectations func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface)
		errorExpected bool
	}{
		{
			name:   "increments distinct counters for a new uri and user",
			update: update,
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statMessages).Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO uris .* IF NOT EXISTS`), update.Uri).Return(q)
				q.EXPECT().MapScanCAS(gomock.Any()).Return(true, nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statUrls).Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO wiki_users .* IF NOT EXISTS`), true, "BotUser").Return(q)
				q.EXPECT().MapScanCAS(gomock.Any()).Return(true, nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statBots).Return(q)
				q.EXPECT().Exec().Return(nil)
			},
			errorExpected: false,
		},
		{
			name:   "skips distinct counters when uri and user already exist",
			update: update,
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statMessages).Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO uris .* IF NOT EXISTS`), update.Uri).Return(q)
				q.EXPECT().MapScanCAS(gomock.Any()).Return(false, nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO wiki_users .* IF NOT EXISTS`), true, "BotUser").Return(q)
				q.EXPECT().MapScanCAS(gomock.Any()).Return(false, nil)
			},
			errorExpected: false,
		},
		{
			name:   "returns error when message counter update fails",
			update: update,
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statMessages).Return(q)
				q.EXPECT().Exec().Return(assert.AnError)
			},
			errorExpected: true,
		},
		{
			name:   "returns error when uri insert fails",
			update: update,
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`UPDATE stats .*`), 1, statMessages).Return(q)
				q.EXPECT().Exec().Return(nil)
				m.EXPECT().Query(test_utils.NewRegexMatcher(`INSERT INTO uris .* IF NOT EXISTS`), update.Uri).Return(q)
				q.EXPECT().MapScanCAS(gomock.Any()).Return(false, assert.AnError)
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
	// expectStat sets up a counter point-read for the given name returning value.
	expectStat := func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface, name string, value int) {
		m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT value FROM stats WHERE name = \?`), name).Return(q)
		q.EXPECT().Scan(gomock.Any()).DoAndReturn(func(arg *int) error {
			*arg = value
			return nil
		})
	}
	tests := []struct {
		name string
		expectations func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface)
		expected models.Statistics
	}{
		{
			name: "returns counter values",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				expectStat(m, q, statMessages, 100)
				expectStat(m, q, statUrls, 50)
				expectStat(m, q, statBots, 30)
				expectStat(m, q, statNonBots, 70)
			},
			expected: models.Statistics{Messages: 100, Urls: 50, Bots: 30, NonBots: 70},
		},
		{
			name: "missing counters report zero",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT value FROM stats WHERE name = \?`), gomock.Any()).Return(q).Times(4)
				q.EXPECT().Scan(gomock.Any()).Return(gocql.ErrNotFound).Times(4)
			},
			expected: models.Statistics{},
		},
		{
			name: "query errors report zero without failing",
			expectations: func(m *mock_stores.MockSessionInterface, q *mock_stores.MockQueryInterface) {
				m.EXPECT().Query(test_utils.NewRegexMatcher(`SELECT value FROM stats WHERE name = \?`), gomock.Any()).Return(q).Times(4)
				q.EXPECT().Scan(gomock.Any()).Return(assert.AnError).Times(4)
			},
			expected: models.Statistics{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := mock_stores.NewMockSessionInterface(ctrl)
			mock_query := mock_stores.NewMockQueryInterface(ctrl)
			tt.expectations(mockSession, mock_query)
			db := &Cassandra{session: mockSession}
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

