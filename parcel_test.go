package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
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

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {

	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	testID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, testID)

	// get
	testParcel, err := store.Get(testID)
	require.NoError(t, err)
	assert.Equal(t, parcel.Address, testParcel.Address)
	assert.Equal(t, parcel.Client, testParcel.Client)
	assert.Equal(t, parcel.CreatedAt, testParcel.CreatedAt)
	assert.Equal(t, parcel.Status, testParcel.Status)
	assert.NotEqual(t, parcel.Number, testParcel.Number)

	// delete
	err = store.Delete(testID)
	require.NoError(t, err)
	_, err = store.Get(testParcel.Number)
	assert.Error(t, err)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	testParcel := getTestParcel()

	// add
	testParcelID, err := store.Add(testParcel)
	require.NoError(t, err)
	require.NotEmpty(t, testParcelID)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(testParcel.Number, newAddress)
	require.NoError(t, err)

	// check
	newTestParcel, err := store.Get(testParcelID)
	require.NoError(t, err)
	require.NotEqual(t, getTestParcel(), newTestParcel)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)

	// set status
	err = store.SetStatus(parcelID, ParcelStatusSent)
	require.NoError(t, err)

	// check
	parcelWithNewStatus, err := store.Get(parcelID)
	require.NoError(t, err)
	assert.NotEqual(t, parcel, parcelWithNewStatus)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		storedParcel, exist := parcelMap[parcel.Number]
		assert.True(t, exist)

		assert.Equal(t, storedParcel, parcel)
	}
}
