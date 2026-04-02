package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-hub/internal/server";"github.com/stockyard-dev/stockyard-hub/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9110"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./hub-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("hub: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Hub — request relay and inspector\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("hub: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
