package db

import (
	"context"
	"input-backend-master-class/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func CreateRandomUser(t *testing.T) Account {
	arg := CreateUserParams{
		Owner:        util.RandomOwner(),
		HashPassword: "secret",
		FullName:     util.RandomOwner(),
		Email:        util.RandomEmail(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	// idとcreated_atは自動生成されるので、0とnilでないことを確認する
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	CreateRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.Owner, account2.Owner)
	require.Equal(t, account1.Balance, account2.Balance)
	require.Equal(t, account1.Currency, account2.Currency)
	require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
}
