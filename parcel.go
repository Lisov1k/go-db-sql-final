package main

import (
	"database/sql"
	"log"
	"time"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	currentTime := time.Now().UTC().Format(time.RFC3339)

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer db.Close()

	res, err := db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (:cl, :status, :address, :curTime)",
		sql.Named("cl", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("curTime", currentTime))
	if err != nil {
		log.Print(err)
		return 0, err
	}

	ID, err := res.LastInsertId()
	if err != nil {
		log.Print(err)
		return 0, err
	}

	return int(ID), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return Parcel{}, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = :number",
		sql.Named("number", number))
	err = row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		log.Print(err)
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var res []Parcel

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = :client",
		sql.Named("client", client))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	for rows.Next() {
		var parcel Parcel

		err = rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		res = append(res, parcel)
	}

	if err := rows.Err(); err != nil {
		log.Print(err)
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = db.Exec("UPDATE parcel SET status = :status WHERE number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = db.Exec("UPDATE parcel SET address = :address WHERE number = :number AND status = 'registered'",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = db.Exec("DELETE FROM parcel WHERE number = :number AND status = 'registered'",
		sql.Named("number", number))
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
