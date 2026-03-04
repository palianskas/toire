package internal

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/palianskas/toire/models"
)

const test_db_path = "/tmp/toire_test.db"

func TestMain(m *testing.M) {
	setup()
	m.Run()
	teardown()
}

func Test_Open_should_openDbWithGivenPath(t *testing.T) {
	_, err := Open(test_db_path)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Init_should_createTables_when_dbEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	err = repo.Init()

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Init_should_doNothing_when_dbInitialized(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	err = repo.Init()

	if err != nil {
		t.Fatal(err)
	}
}

func Test_SaveSeries_should_saveNewEntity_when_IdEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	model := models.Series{
		Name: "name 1",
	}

	err = repo.SaveSeries(&model)

	if err != nil {
		t.Fatal(err)
	}

	if model.Id < 1 {
		t.Fatalf("Expected Id to be set but found %d", model.Id)
	}
}

func Test_SaveSeries_should_updateEntity_when_IdGiven(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}
	model := models.Series{
		Name: "old name",
	}

	err = repo.SaveSeries(&model)

	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	model.Name = "new name"

	err = repo.SaveSeries(&model)

	if err != nil {
		t.Fatal(err)
	}

	var dbmodel *models.Series

	dbmodel, err = repo.GetSeries(model.Id)

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetSeriesMany_should_returnEmpty_when_table_empty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	err = clearTable("series", repo)

	if err != nil {
		t.Fatalf("Failed to clear table. %v", err)
	}

	var series *[]models.Series
	series, err = repo.GetSeriesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*series) != 0 {
		t.Fatalf("Expected 0 rows to be found, but found %d", len(*series))
	}
}

func Test_GetSeriesMany_should_return_fullPage_when_tableFull(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	model := models.Series{
		Name: "name",
	}

	err = repo.SaveSeries(&model)
	if err != nil {
		t.Fatal(err)
	}

	lid := model.Id

	eps := 20
	for i := 1; i < eps; i++ {
		model.Id = -1

		err = repo.SaveSeries(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	rid := model.Id

	var series *[]models.Series
	series, err = repo.GetSeriesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*series) != eps {
		t.Fatalf("Expected %d rows to be found, but found %d", eps, len(*series))
	}
	if rid-lid != eps-1 {
		t.Fatalf("Expected row ids in range (%d:%d), but found (%d:%d)", lid, lid+eps, lid, rid)
	}
}

func Test_GetSeriesMany_should_return_partialPage_when_tableSemiFull(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	model := models.Series{
		Name: "name",
	}

	clearTable("series", repo)

	err = repo.SaveSeries(&model)
	if err != nil {
		t.Fatal(err)
	}

	lid := model.Id

	eps := 10
	for i := 1; i < eps; i++ {
		model.Id = -1

		err = repo.SaveSeries(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	rid := model.Id

	var series *[]models.Series
	series, err = repo.GetSeriesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*series) != eps {
		t.Fatalf("Expected %d rows to be found, but found %d", eps, len(*series))
	}
	if rid-lid != eps-1 {
		t.Fatalf("Expected row ids in range (%d:%d), but found (%d:%d)", lid, lid+eps-1, lid, rid)
	}
}

func Test_SaveSeason_should_saveNewEntity_when_IdEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeriesMany()
	if err != nil {
		t.Fatalf("Failed to load series for relations. Err: %v", err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No series found for relations`)
	}

	model := models.Season{
		Number:   1,
		SeriesId: (*rows)[0].Id,
	}

	err = repo.SaveSeason(&model)

	if err != nil {
		t.Fatal(err)
	}

	if model.Id < 1 {
		t.Fatalf("Expected Id to be set but found %d", model.Id)
	}
}

func Test_SaveSeason_should_updateEntity_when_IdGiven(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeriesMany()
	if err != nil {
		t.Fatalf(`Failed to load series for relations. Err: %v`, err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No series found for relations`)
	}

	model := models.Season{
		Name:     "old name",
		Number:   1,
		SeriesId: (*rows)[0].Id,
	}

	err = repo.SaveSeason(&model)
	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	model.Name = "new name"
	model.Number = 2

	err = repo.SaveSeason(&model)

	if err != nil {
		t.Fatal(err)
	}

	var dbmodel *models.Season

	dbmodel, err = repo.GetSeason(model.Id)

	if err != nil {
		t.Fatal(err)
	}

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetSeasonsMany_should_returnEmpty_when_table_empty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	err = clearTable("seasons", repo)

	if err != nil {
		t.Fatalf("Failed to clear table. %v", err)
	}

	var rows *[]models.Season
	rows, err = repo.GetSeasonsMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*rows) != 0 {
		t.Fatalf("Expected 0 rows to be found, but found %d", len(*rows))
	}
}

func Test_GetSeasonsMany_should_return_allEntities(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeriesMany()
	if err != nil {
		t.Fatalf("Failed to load series for relations. Err: %v", err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No series found for relations`)
	}

	model := models.Season{
		Name:     "name",
		SeriesId: (*rows)[0].Id,
	}

	err = repo.SaveSeason(&model)
	if err != nil {
		t.Fatal(err)
	}

	lid := model.Id

	eps := 20
	for i := 1; i < eps; i++ {
		model.Id = -1

		err = repo.SaveSeason(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	rid := model.Id

	var series *[]models.Season
	series, err = repo.GetSeasonsMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*series) != eps {
		t.Fatalf("Expected %d rows to be found, but found %d", eps, len(*series))
	}
	if rid-lid != eps-1 {
		t.Fatalf("Expected row ids in range (%d:%d), but found (%d:%d)", lid, lid+eps, lid, rid)
	}
}

func Test_GetSeries_should_getEntity_ifIdExists(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	model := models.Series{
		Name: "name 2",
	}

	err = repo.SaveSeries(&model)

	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	var dbmodel *models.Series
	dbmodel, err = repo.GetSeries(model.Id)

	if err != nil {
		t.Fatal(err)
	}

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetSeries_should_fail_ifIdNonExistant(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	_, err = repo.GetSeries(123456789)

	if err == nil {
		t.Fatalf("Expected error but found nil")
	}

	errstr := "sql: Scan error on column index 0, name \"id\": converting NULL to int is unsupported"
	if !strings.Contains(err.Error(), errstr) {
		t.Fatalf("Expected '%v' error, but found '%v'", errstr, err)
	}
}

func Test_GetSeason_should_getEntity_ifIdExists(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeriesMany()
	if err != nil {
		t.Fatalf("Failed to load series for relations. Err: %v", err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No series found for relations`)
	}

	model := models.Season{
		Name:     "name 123",
		Number:   123,
		SeriesId: (*rows)[0].Id,
	}

	err = repo.SaveSeason(&model)

	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	var dbmodel *models.Season
	dbmodel, err = repo.GetSeason(model.Id)

	if err != nil {
		t.Fatal(err)
	}

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetSeason_should_fail_ifIdNonExistant(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	_, err = repo.GetSeason(123456789)

	if err == nil {
		t.Fatalf("Expected error but found nil")
	}

	errstr := "no rows in result set"
	if !strings.Contains(err.Error(), errstr) {
		t.Fatalf("Expected '%v' error, but found '%v'", errstr, err)
	}
}

func Test_SaveEpisode_should_saveNewEntity_when_IdEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeasonsMany()
	if err != nil {
		t.Fatalf(`Failed to load seasons for relations. Err: %v`, err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No seasons found for relations`)
	}

	model := models.Episode{
		Name:     "name 1",
		Number:   1,
		Path:     "path 1",
		SeasonId: (*rows)[0].Id,
	}

	err = repo.SaveEpisode(&model)

	if err != nil {
		t.Fatal(err)
	}

	if model.Id < 1 {
		t.Fatalf("Expected Id to be set but found %d", model.Id)
	}
}

func Test_SaveEpisode_should_updateEntity_when_IdGiven(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeasonsMany()
	if err != nil {
		t.Fatalf(`Failed to load seasons for relations. Err: %v`, err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No seasons found for relations`)
	}

	model := models.Episode{
		Name:     "old name",
		Number:   1,
		Path:     "old path",
		SeasonId: (*rows)[0].Id,
	}

	err = repo.SaveEpisode(&model)

	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	model.Name = "new name"
	model.Number = 2
	model.Path = "new path"

	err = repo.SaveEpisode(&model)

	if err != nil {
		t.Fatal(err)
	}

	var dbmodel *models.Episode
	dbmodel, err = repo.GetEpisode(model.Id)

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetEpisode_should_getEntity_ifIdExists(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeasonsMany()
	if err != nil {
		t.Fatalf(`Failed to load seasons for relations. Err: %v`, err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No seasons found for relations`)
	}

	model := models.Episode{
		Name:     "name 2",
		Number:   2,
		Path:     "path 2",
		SeasonId: (*rows)[0].Id,
	}

	err = repo.SaveEpisode(&model)

	if err != nil {
		t.Fatalf("Failed saving new entity. %v", err)
	}

	var dbmodel *models.Episode
	dbmodel, err = repo.GetEpisode(model.Id)

	if err != nil {
		t.Fatal(err)
	}

	if !dbmodel.Equals(&model) {
		t.Fatalf("Expected dbmodel(%v) to equal updated model(%v), but found mismatch", dbmodel, model)
	}
}

func Test_GetEpisode_should_fail_ifIdNonExistant(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	_, err = repo.GetEpisode(123456789)

	if err == nil {
		t.Fatalf("Expected error but found nil")
	}

	errstr := "no rows in result set"
	if !strings.Contains(err.Error(), errstr) {
		t.Fatalf("Expected '%v' error, but found '%v'", errstr, err)
	}
}

func Test_GetEpisodesMany_should_returnEmpty_when_table_empty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	err = clearTable("episodes", repo)

	if err != nil {
		t.Fatalf("Failed to clear table. %v", err)
	}

	var rows *[]models.Episode
	rows, err = repo.GetEpisodesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*rows) != 0 {
		t.Fatalf("Expected 0 rows to be found, but found %d", len(*rows))
	}
}

func Test_GetEpisodesMany_should_return_fullPage_when_tableFull(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	rows, err := repo.GetSeasonsMany()
	if err != nil {
		t.Fatalf(`Failed to load seasons for relations. Err: %v`, err)
	}
	if len(*rows) < 1 {
		t.Fatal(`No seasons found for relations`)
	}

	model := models.Episode{
		Name:     "name",
		SeasonId: (*rows)[0].Id,
	}

	err = repo.SaveEpisode(&model)
	if err != nil {
		t.Fatal(err)
	}

	lid := model.Id

	eps := 20
	for i := 1; i < eps; i++ {
		model.Id = -1

		err = repo.SaveEpisode(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	rid := model.Id

	var series *[]models.Episode
	series, err = repo.GetEpisodesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*series) != eps {
		t.Fatalf("Expected %d rows to be found, but found %d", eps, len(*series))
	}
	if rid-lid != eps-1 {
		t.Fatalf("Expected row ids in range (%d:%d), but found (%d:%d)", lid, lid+eps, lid, rid)
	}
}

func Test_GetEpisodesMany_should_return_partialPage_when_tableSemiFull(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	seasonRows, err := repo.GetSeasonsMany()
	if err != nil {
		t.Fatalf(`Failed to load seasons for relations. Err: %v`, err)
	}
	if len(*seasonRows) < 1 {
		t.Fatal(`No seasons found for relations`)
	}

	model := models.Episode{
		Name:     "name",
		SeasonId: (*seasonRows)[0].Id,
	}

	clearTable("episodes", repo)

	err = repo.SaveEpisode(&model)
	if err != nil {
		t.Fatal(err)
	}

	lid := model.Id

	eps := 10
	for i := 1; i < eps; i++ {
		model.Id = -1

		err = repo.SaveEpisode(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	rid := model.Id

	var rows *[]models.Episode
	rows, err = repo.GetEpisodesMany()

	if err != nil {
		t.Fatal(err)
	}

	if len(*rows) != eps {
		t.Fatalf("Expected %d rows to be found, but found %d", eps, len(*rows))
	}
	if rid-lid != eps-1 {
		t.Fatalf("Expected row ids in range (%d:%d), but found (%d:%d)", lid, lid+eps-1, lid, rid)
	}
}

func Test_GetTableRowCount_should_returnZero_when_tableEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	clearTable("episodes", repo)

	var count int
	count, err = repo.GetTableRowCount("episodes")
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("Expected row count to be %d, but found %d", 0, count)
	}
}

func Test_GetTableRowCount_should_returnTableRowCount_when_tableNonEmpty(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	clearTable("episodes", repo)

	model := models.Episode{
		Name: "name",
	}

	ec := 10
	for i := 0; i < ec; i++ {
		model.Id = -1

		err = repo.SaveEpisode(&model)
		if err != nil {
			t.Fatal(err)
		}
	}

	var count int
	count, err = repo.GetTableRowCount("episodes")
	if err != nil {
		t.Fatal(err)
	}

	if count != ec {
		t.Fatalf("Expected row count to be %d, but found %d", ec, count)
	}
}

func Test_GetTableRowCount_should_returnError_when_tableNotFound(t *testing.T) {
	repo, err := Open(test_db_path)

	if err != nil {
		t.Fatalf("Failed to open repo %v", err)
	}

	var count int
	count, err = repo.GetTableRowCount("non_existant_table")

	if err == nil {
		t.Fatalf("Expected to receive error but got nil")
	}

	ec := -1
	if count != ec {
		t.Fatalf("Expected to receive count %d but got %d", ec, count)
	}

	ees := "no such table"
	if !strings.Contains(err.Error(), ees) {
		t.Fatalf("Expected to receive error '%v' but got '%v'", ees, err)
	}
}

func setup() {
	os.Create(test_db_path)
}

func teardown() {
	os.RemoveAll(test_db_path)
}

func clearTable(table string, repo *Repo) error {
	_, err := repo.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE true", table))

	return err
}
