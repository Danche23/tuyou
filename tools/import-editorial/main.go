package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	mysqldriver "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"tuyou/internal/model/entity"
	"tuyou/pkg/config"
)

type author struct {
	Key      string `json:"key"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

type post struct {
	Kind       string   `json:"kind"`
	Author     string   `json:"author"`
	Photo      string   `json:"photo"`
	Title      string   `json:"title"`
	Category   int      `json:"category"`
	Tags       []string `json:"tags"`
	Paragraphs []string `json:"paragraphs"`
	Distance   string   `json:"distance"`
	Duration   string   `json:"duration"`
	Difficulty string   `json:"difficulty"`
	Transport  string   `json:"transport"`
}

type contentPack struct {
	Collection string   `json:"collection"`
	Authors    []author `json:"authors"`
	Posts      []post   `json:"posts"`
}

type photoCredit struct {
	Key        string `json:"key"`
	Author     string `json:"author"`
	License    string `json:"license"`
	LicenseURL string `json:"licenseUrl"`
	Source     string `json:"source"`
	File       string `json:"file"`
	SHA256     string `json:"sha256"`
}

func main() {
	manifest := flag.String("manifest", "", "editorial content JSON")
	creditsPath := flag.String("credits", "", "photo attribution JSON")
	photoBase := flag.String("photo-base", "", "published photo URL prefix")
	configPath := flag.String("config", "", "private application config YAML (apply only)")
	apply := flag.Bool("apply", false, "insert data in one transaction")
	flag.Parse()

	if *manifest == "" || *creditsPath == "" || *photoBase == "" {
		fatal(errors.New("--manifest, --credits and --photo-base are required"))
	}
	var pack contentPack
	var credits []photoCredit
	readJSON(*manifest, &pack)
	readJSON(*creditsPath, &credits)
	photoByKey, err := validate(pack, credits, *photoBase)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("validated %d authors, %d posts, %d licensed photos\n", len(pack.Authors), len(pack.Posts), len(credits))
	if !*apply {
		fmt.Println("dry run: no database changes")
		return
	}
	if *configPath == "" {
		fatal(errors.New("--config is required with --apply"))
	}
	if err := checkPublishedPhotos(credits, *photoBase); err != nil {
		fatal(err)
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		fatal(err)
	}
	dbCfg := cfg.Database.MySQL
	dsn := (&mysqldriver.Config{
		User: dbCfg.Username, Passwd: dbCfg.Password, Net: "tcp",
		Addr: fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port), DBName: dbCfg.Database,
		ParseTime: true, Loc: time.Local,
	}).FormatDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fatal(fmt.Errorf("database connection: %w", err))
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		return importPack(tx, pack, photoByKey, *photoBase)
	}); err != nil {
		fatal(err)
	}
	fmt.Println("editorial import committed")
}

func readJSON(path string, out any) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		fatal(err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		fatal(fmt.Errorf("%s: %w", path, err))
	}
}

func validate(pack contentPack, credits []photoCredit, base string) (map[string]photoCredit, error) {
	if pack.Collection == "" || len(pack.Authors) == 0 || len(pack.Posts) == 0 ||
		strings.TrimRight(base, "/") != "https://tuyou.evenstar.love/editorial" {
		return nil, errors.New("invalid collection, empty content or unexpected photo host")
	}
	photoByKey := make(map[string]photoCredit, len(credits))
	for _, credit := range credits {
		if credit.Key == "" || credit.Author == "" || credit.License == "" ||
			!strings.HasPrefix(credit.Source, "https://commons.wikimedia.org/wiki/File:") ||
			!strings.HasPrefix(credit.LicenseURL, "https://") || credit.File != "photos/"+credit.Key+".jpg" || len(credit.SHA256) != 64 {
			return nil, fmt.Errorf("incomplete or unsafe photo credit: %s", credit.Key)
		}
		if _, found := photoByKey[credit.Key]; found {
			return nil, fmt.Errorf("duplicate photo key: %s", credit.Key)
		}
		photoByKey[credit.Key] = credit
	}
	authors := make(map[string]bool, len(pack.Authors))
	for _, a := range pack.Authors {
		if a.Key == "" || len(a.Key) > 12 || authors[a.Key] || utf8.RuneCountInString(a.Nickname) > 20 ||
			utf8.RuneCountInString(a.Bio) > 100 || photoByKey[a.Avatar].Key == "" || !strings.Contains(a.Bio, "途友") {
			return nil, fmt.Errorf("invalid author: %s", a.Key)
		}
		authors[a.Key] = true
	}
	titles := map[string]bool{}
	for _, p := range pack.Posts {
		if (p.Kind != "note" && p.Kind != "guide") || !authors[p.Author] || photoByKey[p.Photo].Key == "" ||
			p.Title == "" || utf8.RuneCountInString(p.Title) > 50 || titles[p.Title] ||
			p.Category < 1 || p.Category > 7 || len(p.Tags) == 0 || len(p.Tags) > 3 || len(p.Paragraphs) < 2 {
			return nil, fmt.Errorf("invalid post: %s", p.Title)
		}
		titles[p.Title] = true
		for _, tag := range p.Tags {
			if tag == "" || utf8.RuneCountInString(tag) > 30 {
				return nil, fmt.Errorf("invalid tag in %s", p.Title)
			}
		}
		if p.Kind == "guide" {
			if p.Distance == "" || p.Duration == "" || p.Transport == "" ||
				utf8.RuneCountInString(p.Distance) > 20 || utf8.RuneCountInString(p.Duration) > 20 ||
				utf8.RuneCountInString(p.Transport) > 20 || difficulty(p.Difficulty) == 0 {
				return nil, fmt.Errorf("invalid route details: %s", p.Title)
			}
		}
		if utf8.RuneCountInString(body(p)) > 2000 {
			return nil, fmt.Errorf("post body too long: %s", p.Title)
		}
	}
	return photoByKey, nil
}

func body(p post) string {
	return strings.Join(p.Paragraphs, "\n\n")
}

func legacyBody(p post, credit photoCredit) string {
	return body(p) + "\n\n配图：" + credit.Author +
		"，Wikimedia Commons，[原图](" + credit.Source + ")，许可：[" + credit.License + "](" + credit.LicenseURL + ")。"
}

func difficulty(value string) int {
	return map[string]int{"入门": 1, "中等": 2, "中高等": 3, "高等": 4}[value]
}

func checkPublishedPhotos(credits []photoCredit, base string) error {
	client := http.Client{Timeout: 20 * time.Second}
	for _, credit := range credits {
		url := strings.TrimRight(base, "/") + "/" + credit.Key + ".jpg"
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("photo unavailable %s: %w", credit.Key, err)
		}
		if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/jpeg") {
			resp.Body.Close()
			return fmt.Errorf("photo unavailable %s: HTTP %d", credit.Key, resp.StatusCode)
		}
		hash := sha256.New()
		written, copyErr := io.Copy(hash, io.LimitReader(resp.Body, (5<<20)+1))
		closeErr := resp.Body.Close()
		if err := errors.Join(copyErr, closeErr); err != nil {
			return fmt.Errorf("photo %s: %w", credit.Key, err)
		}
		if written > 5<<20 || hex.EncodeToString(hash.Sum(nil)) != credit.SHA256 {
			return fmt.Errorf("photo %s: published file differs from attribution manifest", credit.Key)
		}
	}
	return nil
}

func importPack(tx *gorm.DB, pack contentPack, photos map[string]photoCredit, base string) error {
	authorIDs := map[string]uint{}
	for _, a := range pack.Authors {
		phone := "editorial-" + a.Key
		var member entity.Member
		err := tx.Where("phone = ?", phone).First(&member).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				return err
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(secret)), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			member = entity.Member{Phone: phone, Password: string(hash), Nickname: a.Nickname,
				Bio: a.Bio, Avatar: strings.TrimRight(base, "/") + "/" + a.Avatar + ".jpg", Role: 1, Status: 1}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if member.Nickname != a.Nickname || member.Bio != a.Bio {
			return fmt.Errorf("editorial account %s was modified; refusing to reuse", a.Key)
		}
		authorIDs[a.Key] = member.ID
	}
	for _, p := range pack.Posts {
		memberID := authorIDs[p.Author]
		var existingNotes []entity.Note
		if err := tx.Where("title = ?", p.Title).Find(&existingNotes).Error; err != nil {
			return err
		}
		var existingGuides []entity.Strategy
		if err := tx.Where("title = ?", p.Title).Find(&existingGuides).Error; err != nil {
			return err
		}
		url := strings.TrimRight(base, "/") + "/" + p.Photo + ".jpg"
		content := body(p)
		oldContent := legacyBody(p, photos[p.Photo])
		if len(existingNotes)+len(existingGuides) > 0 {
			if p.Kind == "note" && len(existingNotes) == 1 && len(existingGuides) == 0 &&
				existingNotes[0].MemberID == memberID && (existingNotes[0].Content == content || existingNotes[0].Content == oldContent) {
				continue
			}
			if p.Kind == "guide" && len(existingGuides) == 1 && len(existingNotes) == 0 &&
				existingGuides[0].MemberID == memberID && (existingGuides[0].Content == content || existingGuides[0].Content == oldContent) {
				continue
			}
			return fmt.Errorf("title already exists with different data: %s", p.Title)
		}
		if p.Kind == "note" {
			n := entity.Note{MemberID: memberID, Title: p.Title, Content: content, Category: p.Category,
				Status: 1, Heat: 10, PublishTime: time.Now()}
			if err := tx.Create(&n).Error; err != nil {
				return err
			}
			if err := tx.Create(&entity.NoteImage{NoteID: n.ID, URL: url, Sort: 0}).Error; err != nil {
				return err
			}
			if err := attachTags(tx, p.Tags, n.ID, "note"); err != nil {
				return err
			}
		} else {
			g := entity.Strategy{MemberID: memberID, Title: p.Title, Content: content, Category: p.Category,
				Status: 1, Heat: 10, PublishTime: time.Now(), Distance: p.Distance,
				Duration: p.Duration, Difficulty: difficulty(p.Difficulty), Transport: p.Transport}
			if err := tx.Create(&g).Error; err != nil {
				return err
			}
			if err := tx.Create(&entity.StrategyImage{StrategyID: g.ID, URL: url, Sort: 0}).Error; err != nil {
				return err
			}
			if err := attachTags(tx, p.Tags, g.ID, "guide"); err != nil {
				return err
			}
		}
		if err := tx.Model(&entity.Member{}).Where("id = ?", memberID).
			UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error; err != nil {
			return err
		}
	}
	return nil
}

func attachTags(tx *gorm.DB, names []string, contentID uint, kind string) error {
	for _, name := range names {
		var tag entity.Tag
		err := tx.Where("tag_name = ?", name).First(&tag).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tag = entity.Tag{TagName: name, Status: 1}
			if err := tx.Create(&tag).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if tag.Status != 1 {
			return fmt.Errorf("tag disabled: %s", name)
		}
		if kind == "note" {
			if err := tx.Create(&entity.NoteTag{NoteID: contentID, TagID: tag.ID}).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&entity.StrategyTag{StrategyID: contentID, TagID: tag.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
