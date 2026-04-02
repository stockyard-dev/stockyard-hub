package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type App struct{
	ID string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	Icon string `json:"icon"`
	Category string `json:"category"`
	Description string `json:"description"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"hub.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS apps(id TEXT PRIMARY KEY,name TEXT NOT NULL,url TEXT DEFAULT '',icon TEXT DEFAULT '',category TEXT DEFAULT '',description TEXT DEFAULT '',status TEXT DEFAULT 'active',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *App)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO apps(id,name,url,icon,category,description,status,created_at)VALUES(?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.URL,e.Icon,e.Category,e.Description,e.Status,e.CreatedAt);return err}
func(d *DB)Get(id string)*App{var e App;if d.db.QueryRow(`SELECT id,name,url,icon,category,description,status,created_at FROM apps WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.URL,&e.Icon,&e.Category,&e.Description,&e.Status,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]App{rows,_:=d.db.Query(`SELECT id,name,url,icon,category,description,status,created_at FROM apps ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []App;for rows.Next(){var e App;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Icon,&e.Category,&e.Description,&e.Status,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM apps WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM apps`).Scan(&n);return n}
