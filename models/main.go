package models

import "time"

type Series struct {
	Id         int
	Name       string
	SeasonsCnt int
	Path       *string
	CreateDate time.Time

	Seasons *[]Season
}

func (model *Series) Equals(other *Series) bool {
	return model.Id == other.Id &&
		model.Name == other.Name &&
		model.Path == other.Path &&
		model.CreateDate.Equal(other.CreateDate)
}

type Season struct {
	Id         int
	Name       string
	Number     int
	EpisodeCnt int
	Path       *string
	CreateDate time.Time

	SeriesId int
	Series   *Series

	Episodes *[]Episode
}

func (model *Season) Equals(other *Season) bool {
	return model.Id == other.Id &&
		model.Name == other.Name &&
		model.Number == other.Number &&
		model.Path == other.Path &&
		model.CreateDate.Equal(other.CreateDate)
}

type Episode struct {
	Id         int
	Name       string
	Number     int
	Path       string
	IsPlayed   bool
	CreateDate time.Time

	SeasonId int
	Season   *Season
}

func (model *Episode) Equals(other *Episode) bool {
	return model.Id == other.Id &&
		model.Name == other.Name &&
		model.Number == other.Number &&
		model.Path == other.Path &&
		model.IsPlayed == other.IsPlayed &&
		model.CreateDate.Equal(other.CreateDate)
}
