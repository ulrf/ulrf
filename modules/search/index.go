package search

import (
	"github.com/fatih/color"
	"github.com/ulrf/ulrf/models"
)

func IndexTitle(ogrn string, data models.ForIndex) error {
	ix := indexes[indexTitleName]
	e := ix.Index(ogrn, data)
	if e != nil {
		return e
	}

	return nil
}

func IndexTitleBatch(os []models.Org) error {
	ix := indexes[indexTitleName]
	batch := ix.NewBatch()

	for _, o := range os {
		color.Green("indexing %s", o.ShortName)
		batch.Index(o.OGRN, o.ForIndex())
	}

	e := ix.Batch(batch)
	if e != nil {
		return e
	}

	return nil
}

func IndexCity(ogrn string, data models.ForCityIndex) error {
	ix := indexes[indexCityName]
	e := ix.Index(ogrn, data)
	if e != nil {
		return e
	}

	return nil
}

func IndexCityBatch(city string, os []models.Org) error {
	ix := indexes[indexCityName]
	batch := ix.NewBatch()

	for _, o := range os {
		o.City = city
		color.Green("indexing %s", o.City)
		batch.Index(o.OGRN, o.ForCityIndex())
		o.City = ""
		o.OGRN = ""
	}

	e := ix.Batch(batch)
	if e != nil {
		return e
	}

	return nil
}

func IndexOkved(ogrn string, data models.ForOKVEDIndex) error {
	ix := indexes[indexEkvdName]
	e := ix.Index(ogrn, data)
	if e != nil {
		return e
	}

	return nil
}

func IndexOkvedBatch(okved string, os []models.Org) error {
	ix := indexes[indexEkvdName]
	batch := ix.NewBatch()

	for _, o := range os {
		o.OKVED = okved
		color.Green("indexing %s", o.OKVED)
		batch.Index(o.OGRN, o.ForOKVEDIndex())
		o.OKVED = ""
		o.OGRN = ""
		o.OKVEDS = o.OKVEDS[:0]
	}

	e := ix.Batch(batch)
	if e != nil {
		return e
	}

	return nil
}
