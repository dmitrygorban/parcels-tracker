package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// setupStore создает подключение к БД и возвращает store
func setupStore(t *testing.T) ParcelStore {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "No db connection")

	t.Cleanup(func() {
		db.Close()
	})

	return NewParcelStore(db)
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	store := setupStore(t)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err, fmt.Sprintf("Errow adding parcel: %s\n", err))
	require.NotEmpty(t, number, "id is empty")
	parcel.Number = number

	// get
	p, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, parcel, p)

	// delete
	err = store.Delete(p.Number)
	require.NoError(t, err)

	_, err = store.Get(p.Number)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	store := setupStore(t)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err)

	// check
	p, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)
}

// // TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	store := setupStore(t)
	parcel := getTestParcel()
	nextStatus := ParcelStatusDelivered
	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, number)

	// set status
	err = store.SetStatus(number, nextStatus)
	require.NoError(t, err)

	// check
	p, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, nextStatus, p.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	store := setupStore(t)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		storedParcel, exists := parcelMap[parcel.Number]
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		require.True(t, exists)
		// убедитесь, что значения полей полученных посылок заполнены верно
		require.Equal(t, storedParcel, parcel)
	}
}
