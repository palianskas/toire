package internal

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/palianskas/toire/models"
)

type Tables struct {
	Series   string
	Seasons  string
	Episodes string
}

type Repo struct {
	db *sql.DB
	Tables
}

const SERIES_TABLE = "series"
const SEASONS_TABLE = "seasons"
const EPISODES_TABLE = "episodes"

type stateKeys struct {
	LastPlayedEpId string
}

var StateKeys = stateKeys{
	LastPlayedEpId: "last_played_ep_id",
}

var instance *Repo = nil

func Open(path string) (*Repo, error) {
	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return nil, err
	}

	repo := Repo{
		db: db,
		Tables: Tables{
			Series:   SERIES_TABLE,
			Seasons:  SEASONS_TABLE,
			Episodes: EPISODES_TABLE,
		},
	}

	return &repo, err
}

func (repo *Repo) Init() error {
	query := `
CREATE TABLE IF NOT EXISTS series (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	path TEXT NULL,
	create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
CREATE TABLE IF NOT EXISTS seasons (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT DEFAULT "",
	number INTEGER,
	path TEXT NULL,
	create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	series_id INTEGER,
	FOREIGN KEY(series_id) REFERENCES series (id) ON DELETE CASCADE
	);
CREATE TABLE IF NOT EXISTS episodes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT DEFAULT "",
	number INTEGER,
	path TEXT,
	is_played BOOL DEFAULT FALSE,
	create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	season_id INTEGER,
	FOREIGN KEY(season_id) REFERENCES seasons (id) ON DELETE CASCADE
	);

CREATE TABLE IF NOT EXISTS state (
	key VARCHAR(64) PRIMARY KEY,
	value VARCHAR(256)
	); 
`

	_, err := repo.db.Exec(query)

	return err
}

func GetInstance(path string) (*Repo, error) {
	if instance != nil {
		return instance, nil
	}

	repo, err := Open(path)

	if err != nil {
		return nil, err
	}

	err = repo.Init()

	if err != nil {
		return nil, err
	}

	instance = repo

	return instance, nil
}

func (repo *Repo) SaveSeries(model *models.Series) error {
	var err error

	if model.Id < 1 {
		err = repo.createSeries(model)
	} else {
		err = repo.updateSeries(model)
	}

	return err
}

func (repo *Repo) DeleteSeries(id int) error {
	_, err := repo.db.Exec("DELETE FROM series WHERE id = ?", id)

	return err
}

func (repo *Repo) GetSeries(id int) (*models.Series, error) {
	query := `
        SELECT series.id, series.name, series.path, series.create_date, COUNT(seasons.id) as season_cnt 
        FROM series
        LEFT JOIN seasons 
			ON (series.id = seasons.series_id)
        WHERE series.id = ? 
    `

	row := repo.db.QueryRow(query, id)

	var model models.Series
	err := row.Scan(&model.Id, &model.Name, &model.Path, &model.CreateDate, &model.SeasonsCnt)

	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (repo *Repo) GetSeriesMany() (*[]models.Series, error) {
	query := `
        SELECT series.id, series.name, series.path, series.create_date, COUNT(seasons.id) as season_cnt
        FROM series 
        LEFT JOIN seasons 
            ON (seasons.series_id = series.id) 
        GROUP BY series.id 
`

	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, err
	}

	res := make([]models.Series, 0)

	for rows.Next() {
		var model models.Series

		err = rows.Scan(&model.Id, &model.Name, &model.Path, &model.CreateDate, &model.SeasonsCnt)

		if err != nil {
			return nil, err
		}

		res = append(res, model)
	}

	return &res, nil
}

func (repo *Repo) SaveSeason(model *models.Season) error {
	var err error

	if model.Id < 1 {
		err = repo.createSeason(model)
	} else {
		err = repo.updateSeason(model)
	}

	return err
}

func (repo *Repo) DeleteSeason(id int) error {
	_, err := repo.db.Exec("DELETE FROM seasons WHERE id = ?", id)

	return err
}

func (repo *Repo) GetSeason(id int) (*models.Season, error) {
	query := `
        SELECT 
			seasons.id, 
			seasons.number, 
			seasons.name, 
			seasons.path, 
			seasons.create_date,
			COUNT(episodes.id) as ep_cnt,
			series.id, 
			series.name,
			series.path,
			series.create_date
        FROM seasons
        LEFT JOIN episodes ON (seasons.id = episodes.season_id)
        LEFT JOIN series ON (seasons.series_id = series.id)
        WHERE seasons.id = ?
		GROUP BY seasons.id
    `

	row := repo.db.QueryRow(query, id)

	var season models.Season
	var series models.Series

	err := row.Scan(
		&season.Id,
		&season.Number,
		&season.Name,
		&season.Path,
		&season.CreateDate,
		&season.EpisodeCnt,
		&series.Id,
		&series.Name,
		&series.Path,
		&series.CreateDate,
	)

	if err != nil {
		return nil, err
	}

	season.SeriesId = series.Id
	season.Series = &series

	return &season, nil
}

func (repo *Repo) GetSeasonsMany() (*[]models.Season, error) {
	query := `
        SELECT 
			seasons.id, 
			seasons.name, 
			seasons.number, 
			seasons.path, 
			seasons.create_date,
			COUNT(episodes.id) as episode_cnt,
			series.id,
			series.name,
			series.path,
			series.create_date
        FROM seasons
		LEFT JOIN series
			ON (series.id = seasons.series_id)
        LEFT JOIN episodes
            ON (episodes.season_id = seasons.id) 
        GROUP BY seasons.id
`
	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, err
	}

	res := make([]models.Season, 0)

	for rows.Next() {
		var model models.Season
		var series models.Series

		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Number,
			&model.Path,
			&model.CreateDate,
			&model.EpisodeCnt,
			&series.Id,
			&series.Name,
			&series.Path,
			&series.CreateDate,
		)

		model.Series = &series

		if err != nil {
			return nil, err
		}

		res = append(res, model)
	}

	return &res, nil

}

func (repo *Repo) GetSeasonsBySeries(seriesId int) (*[]models.Season, error) {
	query := `
        SELECT 
			seasons.id, 
			seasons.name, 
			seasons.path, 
			seasons.number, 
			seasons.create_date,
			COUNT(episodes.id) as episode_cnt,
			series.id,
			series.name,
			series.path,
			series.create_date
        FROM seasons
		LEFT JOIN series
			ON (series.id = seasons.series_id)
        LEFT JOIN episodes
            ON (episodes.season_id = seasons.id) 
		WHERE seasons.series_id = ?
        GROUP BY seasons.id
`

	rows, err := repo.db.Query(query, seriesId)

	if err != nil {
		return nil, err
	}

	res := make([]models.Season, 0)

	var series models.Series

	for rows.Next() {
		var model models.Season

		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Path,
			&model.Number,
			&model.CreateDate,
			&model.EpisodeCnt,
			&series.Id,
			&series.Name,
			&series.Path,
			&series.CreateDate,
		)

		model.Series = &series

		if err != nil {
			return nil, err
		}

		res = append(res, model)
	}

	return &res, nil
}

func (repo *Repo) SaveEpisode(model *models.Episode) error {
	var err error

	if model.Id < 1 {
		err = repo.createEpisode(model)
	} else {
		err = repo.updateEpisode(model)
	}

	return err
}

func (repo *Repo) DeleteEpisode(id int) error {
	_, err := repo.db.Exec("DELETE FROM episodes WHERE id = ?", id)

	return err
}

func (repo *Repo) GetEpisode(id int) (*models.Episode, error) {
	row := repo.db.QueryRow("SELECT id, name, number, path, is_played, create_date, season_id FROM episodes WHERE id IN (?) LIMIT 1", id)

	var model models.Episode

	err := row.Scan(&model.Id, &model.Name, &model.Number, &model.Path, &model.IsPlayed, &model.CreateDate, &model.SeasonId)

	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (repo *Repo) GetEpisodesMany() (*[]models.Episode, error) {
	query := `
	SELECT 
		episodes.id, 
		episodes.name, 
		episodes.number, 
		episodes.path, 
		episodes.is_played, 
		episodes.create_date,
		seasons.id,
		seasons.name,
		seasons.path,
		seasons.number,
		seasons.create_date,
		series.id,
		series.name,
		series.path,
		series.create_date
	FROM episodes
	LEFT JOIN seasons
		ON (seasons.id = episodes.season_id)
	LEFT JOIN series
		ON (series.id = seasons.series_id)
	`
	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, err
	}

	res := make([]models.Episode, 0)

	for rows.Next() {
		var model models.Episode
		var season models.Season
		var series models.Series

		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Number,
			&model.Path,
			&model.IsPlayed,
			*&model.CreateDate,
			&season.Id,
			&season.Name,
			&season.Path,
			&season.CreateDate,
			&season.Number,
			&series.Id,
			&series.Name,
			&series.Path,
			&series.CreateDate,
		)

		if err != nil {
			return nil, err
		}

		model.Season = &season
		season.Series = &series
		res = append(res, model)
	}

	return &res, nil
}

func (repo *Repo) GetEpisodesBySeason(seasonId int) (*[]models.Episode, error) {
	query := `
	SELECT 
		episodes.id, 
		episodes.name, 
		episodes.number, 
		episodes.path,
		episodes.is_played,
		episodes.create_date,
		seasons.id,
		seasons.name,
		seasons.path,
		seasons.number,
		seasons.create_date,
		series.id,
		series.name,
		series.path,
		series.create_date
	FROM episodes
	LEFT JOIN seasons
		ON (seasons.id = episodes.season_id)
	LEFT JOIN series
		ON (series.id = seasons.series_id)
	WHERE seasons.id = ?
	`

	rows, err := repo.db.Query(query, seasonId)

	if err != nil {
		return nil, err
	}

	res := make([]models.Episode, 0)
	var season models.Season
	var series models.Series

	for rows.Next() {
		var model models.Episode

		err = rows.Scan(
			&model.Id,
			&model.Name,
			&model.Number,
			&model.Path,
			&model.IsPlayed,
			&model.CreateDate,
			&season.Id,
			&season.Name,
			&season.Path,
			&season.Number,
			&season.CreateDate,
			&series.Id,
			&series.Name,
			&series.Path,
			&series.CreateDate,
		)

		if err != nil {
			return nil, err
		}

		model.Season = &season
		season.Series = &series
		res = append(res, model)
	}

	return &res, nil
}

func (repo *Repo) GetTableRowCount(table string) (int, error) {
	row := repo.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))

	var count int
	err := row.Scan(&count)

	if err != nil {
		return -1, err
	}

	return count, nil
}

func (repo *Repo) GetLastPlayedEpisodeId() (int, error) {
	row := repo.db.QueryRow(`SELECT value FROM state WHERE key = ? LIMIT 1`, StateKeys.LastPlayedEpId)

	var value string
	err := row.Scan(&value)

	if err != nil {
		return -1, err
	}

	epId, err := strconv.Atoi(value)

	if err != nil {
		return -2, err
	}

	return epId, nil
}

func (repo *Repo) UpdateLastPlayedEpisodeId(episodeId int) error {
	query := `
	INSERT INTO state(key,value) VALUES (?, ?) 
	ON CONFLICT(key) DO UPDATE SET value = ?
	`

	value := fmt.Sprintf("%d", episodeId)

	_, err := repo.db.Exec(query, StateKeys.LastPlayedEpId, value, value)

	return err
}

func (repo *Repo) createSeries(model *models.Series) error {
	m, err := repo.db.Exec("INSERT INTO series(name, path) VALUES (?, ?)", model.Name, model.Path)

	var id int64
	id, err = m.LastInsertId()

	if err != nil {
		return err
	}

	model.Id = int(id)

	return err
}

func (repo *Repo) updateSeries(model *models.Series) error {
	_, err := repo.db.Exec("UPDATE series SET name=?, path=? WHERE id=?", model.Name, model.Path, model.Id)

	return err
}

func (repo *Repo) createSeason(model *models.Season) error {
	m, err := repo.db.Exec(`INSERT INTO seasons(name, number, series_id) VALUES (?, ?, ?)`, model.Name, model.Number, model.SeriesId)

	if err != nil {
		return err
	}

	var id int64
	id, err = m.LastInsertId()

	if err != nil {
		return err
	}

	model.Id = int(id)

	return err
}

func (repo *Repo) updateSeason(model *models.Season) error {
	_, err := repo.db.Exec("UPDATE seasons SET name=?, number=?, path=? WHERE id=?", model.Name, model.Number, *model.Path, model.Id)

	return err
}

func (repo *Repo) createEpisode(model *models.Episode) error {
	m, err := repo.db.Exec("INSERT INTO episodes(name, number, path, is_played, season_id) VALUES (?, ?, ?, ?, ?)", model.Name, model.Number, model.Path, model.IsPlayed, model.SeasonId)

	if err != nil {
		return err
	}

	var id int64
	id, err = m.LastInsertId()

	if err != nil {
		return err
	}

	model.Id = int(id)

	return err
}

func (repo *Repo) updateEpisode(model *models.Episode) error {
	_, err := repo.db.Exec("UPDATE episodes SET name=?, number=?, path=?, is_played=? WHERE id=?", model.Name, model.Number, model.Path, model.IsPlayed, model.Id)

	return err
}
