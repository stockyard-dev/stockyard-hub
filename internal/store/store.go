package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Service struct {
	ID string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	Type string `json:"type"`
	Owner string `json:"owner"`
	Version string `json:"version"`
	Status string `json:"status"`
	HealthURL string `json:"health_url"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"hub.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS services(id TEXT PRIMARY KEY,name TEXT NOT NULL,url TEXT DEFAULT '',type TEXT DEFAULT 'api',owner TEXT DEFAULT '',version TEXT DEFAULT '',status TEXT DEFAULT 'active',health_url TEXT DEFAULT '',notes TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Service)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO services(id,name,url,type,owner,version,status,health_url,notes,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.URL,e.Type,e.Owner,e.Version,e.Status,e.HealthURL,e.Notes,e.CreatedAt);return err}
func(d *DB)Get(id string)*Service{var e Service;if d.db.QueryRow(`SELECT id,name,url,type,owner,version,status,health_url,notes,created_at FROM services WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.URL,&e.Type,&e.Owner,&e.Version,&e.Status,&e.HealthURL,&e.Notes,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Service{rows,_:=d.db.Query(`SELECT id,name,url,type,owner,version,status,health_url,notes,created_at FROM services ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Service;for rows.Next(){var e Service;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Type,&e.Owner,&e.Version,&e.Status,&e.HealthURL,&e.Notes,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Service)error{_,err:=d.db.Exec(`UPDATE services SET name=?,url=?,type=?,owner=?,version=?,status=?,health_url=?,notes=? WHERE id=?`,e.Name,e.URL,e.Type,e.Owner,e.Version,e.Status,e.HealthURL,e.Notes,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM services WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM services`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Service{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ?)"
        args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["type"];ok&&v!=""{where+=" AND type=?";args=append(args,v)}
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,url,type,owner,version,status,health_url,notes,created_at FROM services WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Service;for rows.Next(){var e Service;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Type,&e.Owner,&e.Version,&e.Status,&e.HealthURL,&e.Notes,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM services GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
