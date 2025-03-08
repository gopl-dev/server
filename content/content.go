package content

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/ghodss/yaml"
	"github.com/go-git/go-git/v5"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/xeipuuv/gojsonschema"
)

type Data struct {
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Preview     string     `yaml:"preview"`
	Links       []DataLink `yaml:"links"`
	Topics      []string   `yaml:"topics"`
}

type DataLink struct {
	Icon string `yaml:"icon"`
	Text string `yaml:"text"`
	Addr string `yaml:"addr"`
}

// type of data is added as topic
// but some exceptions exist
var autoTopicExceptions = []string{
	// definition is special kind of data that describes topic
	// and returned if only one topic selected
	// so "definition" must be hidden, as it is makes no sense
	"definition",
}

var pluralizer = pluralize.NewClient()

func ImportFromGitHub(ctx context.Context, repo ds.GitHubRepo) (err error) {
	startAt := time.Now()

	var (
		//topics    []string
		dataTypes []string
		dataCount int
	)

	defer func() {
		repo.ImportedAt = app.TimeNowPtr()

		if err != nil {
			repo.ImportStatus = ds.ImportFailed
			repo.ImportLog = err.Error()
		} else {
			repo.ImportStatus = ds.ImportSuccess
			dur := time.Since(startAt).String()
			// TODO correct plural/singular form for counters
			var importLog string
			if len(dataTypes) == 1 {
				typeName := dataTypes[0]
				importLog = fmt.Sprintf("%d %s imported in %s", dataCount, pluralizer.Plural(typeName), dur)
			} else {
				importLog = fmt.Sprintf("%d entities of %d types is imported in %s", dataCount, len(dataTypes), dur)
			}
			repo.ImportLog = importLog
		}

		err2 := service.UpdateGitHubRepo(&repo)
		if err2 != nil {
			log.Fatal(err2)
		}
	}()

	// make dir to clone to
	conf := app.Config()
	cloneTo := path.Join(conf.Content.LocalDir, repo.Path)
	err = os.RemoveAll(cloneTo) // If the path does not exist RemoveAll  returns nil error
	if err != nil {
		err = fmt.Errorf("os.RemoveAll: %v", err)
		return
	}
	_ = os.MkdirAll(cloneTo, 0755)

	// clone
	_, err = git.PlainClone(cloneTo, false, &git.CloneOptions{
		URL:               repo.CloneURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		err = fmt.Errorf("git clone: %v", err)
		return
	}

	// validate
	_, err = ValidateJSONSchema(cloneTo)
	if err != nil {
		err = fmt.Errorf("validate json schema : %v", err)
		return
	}

	// import to DB
	err = ImportContentFromDir(ctx, cloneTo, repo.ID)
	if err != nil {
		err = fmt.Errorf("ImportContentFromDir: %v", err)
		return
	}

	//err = database.ORM().
	//	Model(&[]model.Topic{}).
	//	Column("name").
	//	Where("repo_id = ?", repo.ID).
	//	Order("name").
	//	Select(&topics)
	//if err != nil {
	//	err = errors.Wrap(err, "get topics")
	//	return
	//}
	//
	//err = database.ORM().
	//	Model(&[]model.Entity{}).
	//	Column("type").
	//	Where("repo_id = ?", repo.ID).
	//	Order("type").
	//	Group("type").
	//	Select(&dataTypes)
	//if err != nil {
	//	err = errors.Wrap(err, "get data types")
	//	return
	//}
	//
	//dataCount, err = database.ORM().
	//	Model(&[]model.Entity{}).
	//	Where("repo_id = ?", repo.ID).
	//	Count()
	//if err != nil {
	//	err = errors.Wrap(err, "get data count")
	//	return
	//}

	return nil
}

// ImportContentFromDir imports data from local path
// TODO partial import
func ImportContentFromDir(ctx context.Context, dir string, repoID int64) (err error) {
	// Mark all data as deleted,
	// and while importing restore existing entities
	// Entities that still marked as deleted after import should be deleted for real
	//_, err = app.DB().
	//	Exec("UPDATE entities SET deleted_at=NOW() WHERE repo_id = $1", repoID)
	//	return
	//}

	// delete topics
	//_, err = database.ORM().
	//	Exec("DELETE FROM entity_topics WHERE topic_id IN (SELECT id FROM topics WHERE repo_id = ?)", repoID)
	//if err != nil {
	//	return
	//}
	//_, err = database.ORM().
	//	Exec("DELETE FROM topics WHERE repo_id = ?", repoID)
	//if err != nil {
	//	return
	//}

	err = importDataFromDir(dir, repoID)
	if err != nil {
		return
	}

	// TODO: since collections introduced it is not good to hard delete entities,
	//  because that will be confusing when entity from collection simply disappear.
	//  Maybe if entity is in any collection then keep it marked as deleted and display it as deleted when it viewed in collection
	//  Or implement notification system and inform everyone (that have deleted entity in their collection) about deleted entity
	// 	Or make public changelog, so it will be clear to everyone what happened
	//_, err = database.ORM().
	//	Exec("DELETE FROM entities WHERE deleted_at IS NOT NULL AND repo_id = ?", repoID)
	//if err != nil {
	//	return
	//}

	return
}

func importDataFromDir(dir string, repoID int64) (err error) {
	err = filepath.Walk(dir, func(path string, f os.FileInfo, wErr error) (err error) {
		if wErr != nil {
			return wErr
		}

		// skip dirs, samples, schemas and non-yaml files
		if !IsContentFile(f) {
			return
		}

		nameParts := strings.Split(f.Name(), ".")
		var eType string
		if len(nameParts) > 2 {
			eType = nameParts[len(nameParts)-2]
		}

		fData, err := ioutil.ReadFile(path)
		if err != nil {
			return
		}

		dataEl := Data{}
		err = yaml.Unmarshal(fData, &dataEl)
		if err != nil {
			return
		}

		extWithType := eType + YAMLExt
		if extWithType[0] != '.' {
			extWithType = "." + extWithType
		}

		var entityData ds.EntityData
		err = yaml.Unmarshal(fData, &entityData)
		if err != nil {
			return
		}

		// prepend type of data to topic
		prependTopics := make([]string, 0)
		if eType != "" && useTypeAsTopic(eType) {
			prependTopics = append(prependTopics, eType)
		}
		if eType == "definition" {
			prependTopics = append(prependTopics, nameParts[len(nameParts)-3])
		}

		if len(prependTopics) > 0 {
			entityData = addTypeToTopics(entityData, prependTopics...)
			dataEl.Topics = append([]string{eType}, dataEl.Topics...)
		}

		mEntity := ds.Entity{
			RepoID:    repoID,
			Path:      app.RelativeFilePath(dir, path),
			Title:     dataEl.Title,
			Type:      eType,
			Data:      entityData,
			CreatedAt: time.Now(),
		}

		err = service.CreateOrUpdateEntity(&mEntity)
		if err != nil {
			return
		}

		for _, name := range dataEl.Topics {
			mTopic := ds.Topic{
				Name:   name,
				RepoID: repoID,
			}
			err := service.FirstOrCreateTopic(&mTopic) // TODO keep cache in memory?
			if err != nil {
				return err
			}
			err = service.CreateEntityTopic(ds.EntityTopic{
				EntityID: mEntity.ID,
				TopicID:  mTopic.ID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func addTypeToTopics(data ds.EntityData, t ...string) ds.EntityData {
	newTopics := make([]interface{}, len(t))
	for i, v := range t {
		newTopics[i] = v
	}
	dt, ok := data["topics"]
	if !ok {
		data["topics"] = newTopics
		return data
	}

	topics, ok := dt.([]interface{})
	if !ok {
		data["topics"] = newTopics
		return data
	}

	data["topics"] = append(newTopics, topics...)
	return data
}

func useTypeAsTopic(t string) bool {
	for _, v := range autoTopicExceptions {
		if v == t {
			return false
		}
	}

	return true
}

type Type string

const (
	GenericType Type = "generic"
)

const (
	YAMLExt   = ".yaml"
	sampleExt = ".sample.yaml"
	schemaExt = ".schema.yaml"
)

func IsSampleFile(fName string) bool {
	return strings.HasSuffix(fName, sampleExt)
}

func IsSchemaFile(fName string) bool {
	return strings.HasSuffix(fName, schemaExt)
}

func IsYamlFile(fName string) bool {
	return strings.HasSuffix(fName, YAMLExt)
}

func IsContentFile(f os.FileInfo) (ok bool) {
	if f.IsDir() {
		return false
	}
	if IsSchemaFile(f.Name()) {
		return false
	}
	if IsSampleFile(f.Name()) {
		return false
	}
	if IsYamlFile(f.Name()) {
		return true
	}

	return false
}

func JSONBytesFromYAMLFile(fPath string) (data []byte, err error) {
	yamlData, err := os.ReadFile(fPath)
	if err != nil {
		return
	}

	data, err = yaml.YAMLToJSON(yamlData)
	return
}

func TypeFromFilename(fPath string) (t Type) {
	_, fName := path.Split(fPath)
	nameParts := strings.Split(fName, ".")
	if len(nameParts) > 2 {
		t = Type(nameParts[len(nameParts)-2])
	}

	if t == "" {
		t = GenericType
	}

	return
}

func TypeFromSchemaFilename(fPath string) (t Type) {
	_, fName := filepath.Split(fPath)
	nameParts := strings.Split(fName, ".")
	if len(nameParts) > 2 {
		t = Type(nameParts[0])
	}

	if t == "" {
		t = GenericType
	}

	return
}

type Schema struct {
	// I need Path only to display informative error messages
	Path   string
	Schema *gojsonschema.Schema
}

type File struct {
	// I need Path only to display informative error messages
	Path string
	Data []byte
}

type schemas map[Type]Schema
type filesByType map[Type][]File

type ValidateResult struct {
	SchemaCount     int
	DataCount       int
	DataCountByType map[Type]int
}

type ErrStack []string

func (e *ErrStack) Add(err error, wrapOpt ...string) {
	if len(wrapOpt) > 0 {
		err = errors.New(wrapOpt[0] + ": " + err.Error())
	}

	*e = append(*e, err.Error())
}
func (e ErrStack) Error() string {
	return strings.Join(e, "\n")
}

// ValidateJSONSchema validates each YAML doc against JSON schema of it's type
func ValidateJSONSchema(dirPath string) (resp ValidateResult, err error) {
	_, err = os.Stat(dirPath)
	if err != nil {
		return
	}

	// add trailing slash, so base name of file that not have it
	// when displaying errors and messages
	if !strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath += string(filepath.Separator)
	}

	resp = ValidateResult{
		SchemaCount:     0,
		DataCount:       0,
		DataCountByType: map[Type]int{},
	}

	schemasRepo := schemas{}
	dataRepo := filesByType{}
	errs := ErrStack{}

	// collect schemas along with data (to not walk dirs seconds time)
	err = filepath.Walk(dirPath, func(path string, f os.FileInfo, wErr error) (err error) {
		if wErr != nil {
			return wErr
		}

		if IsSchemaFile(f.Name()) {
			err = registerSchema(dirPath, path, schemasRepo)
			if err != nil {
				errs.Add(err, "register schema: "+app.RelativeFilePath(dirPath, path))
				return nil
			}
			resp.SchemaCount++
			return nil
		}

		// skip dirs, samples and non-yaml files
		if !IsContentFile(f) {
			return
		}

		resp.DataCount++
		jsonBytes, err := JSONBytesFromYAMLFile(path)
		if err != nil {
			errs.Add(err, app.RelativeFilePath(dirPath, path))
			return nil
		}

		t := TypeFromFilename(app.RelativeFilePath(dirPath, path))
		_, ok := resp.DataCountByType[t]
		if !ok {
			resp.DataCountByType[t] = 0
		}
		resp.DataCountByType[t]++
		files, ok := dataRepo[t]
		if !ok {
			files = []File{}
		}
		files = append(files, File{
			Path: path,
			Data: jsonBytes,
		})
		dataRepo[t] = files

		return nil
	})

	if len(errs) > 0 {
		err = errs
		return
	}

	// iterate over collected data and validate each
	for t, files := range dataRepo {
		for _, v := range files {
			schema, ok := schemasRepo[t]
			if !ok {
				errs.Add(fmt.Errorf("schema of type '%s' is not exists (source '%s')", t, app.RelativeFilePath(dirPath, v.Path)))
				break
			}

			loader := gojsonschema.NewBytesLoader(v.Data)
			result, err := schema.Schema.Validate(loader)
			if err != nil {
				errs.Add(err, "validate "+app.RelativeFilePath(dirPath, v.Path))
				continue
			}

			if !result.Valid() {
				errs.Add(errors.New("validation failed for " + app.RelativeFilePath(dirPath, v.Path)))
				for _, e := range result.Errors() {
					errs.Add(errors.New("\t - " + e.String()))
				}
			}
		}
	}

	if len(errs) > 0 {
		err = errs
	}

	return
}

// registerSchema loads YAML schema from fPath converts it to JSON and adds it to the repo
// returns error if schema of given type already registered
func registerSchema(dirPath, filePath string, repo schemas) (err error) {
	t := TypeFromSchemaFilename(filePath)

	// check if this schema type already registered
	existing, ok := repo[t]
	if ok {
		err = fmt.Errorf(
			"schema type '%s' from '%s' already registered in '%s'",
			t, app.RelativeFilePath(dirPath, filePath), app.RelativeFilePath(dirPath, existing.Path),
		)
		return
	}

	jsonBytes, err := JSONBytesFromYAMLFile(filePath)
	if err != nil {
		return
	}

	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader(jsonBytes)
	schema, err := sl.Compile(loader)
	if err != nil {
		return
	}

	repo[t] = Schema{
		Path:   filePath,
		Schema: schema,
	}

	return
}
