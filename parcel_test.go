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

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	testID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, testID)
	assert.NotNil(t, testID)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	testParcel, err := store.Get(testID)
	require.NoError(t, err)
	assert.Equal(t, parcel.Address, testParcel.Address)
	assert.Equal(t, parcel.Client, testParcel.Client)
	assert.Equal(t, parcel.CreatedAt, testParcel.CreatedAt)
	assert.Equal(t, parcel.Status, testParcel.Status)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
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
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	testParcelID, err := store.Add(testParcel)
	require.NoError(t, err)
	require.NotEmpty(t, testParcelID)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(testParcel.Number, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	newTestParcel, err := store.Get(testParcelID)
	require.NoError(t, err)
	require.NotEqual(t, getTestParcel(), newTestParcel)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcelID)
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set status
	err = store.SetStatus(parcelID, ParcelStatusSent)
	require.NoError(t, err)
	// обновите статус, убедитесь в отсутствии ошибки

	// check
	parcelWithNewStatus, err := store.Get(parcelID)
	require.NoError(t, err)
	assert.NotEqual(t, parcel, parcelWithNewStatus)
	// получите добавленную посылку и убедитесь, что статус обновился
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
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Equal(t, len(parcels), len(storedParcels))
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		storedParcel, exist := parcelMap[parcel.Number]
		require.True(t, exist)
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Equal(t, storedParcel.Client, parcel.Client)
		assert.Equal(t, storedParcel.Address, parcel.Address)
		assert.Equal(t, storedParcel.Status, parcel.Status)
		assert.Equal(t, storedParcel.CreatedAt, parcel.CreatedAt)
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
